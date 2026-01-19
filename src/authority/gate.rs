// ============================================================
// FILE: src/authority/gate.rs
// ============================================================

use super::types::*;
use super::errors::*;
use super::receipt::ReceiptMint;

// -----------------------------------------------------------------------------
// Injected traits (implemented elsewhere) — authority boundaries
// -----------------------------------------------------------------------------

pub trait FreezeStateReader {
    fn is_frozen(&self, domain: &DomainRef, scope: &Scope) -> Result<bool, AuthorityError>;
}

pub trait FreezeStateWriter {
    fn apply_freeze_op(
        &mut self,
        domain: &DomainRef,
        scope: &Scope,
        op: FreezeOp,
        ttl_hint_seconds: Option<u64>,
    ) -> Result<String, AuthorityError>;
}

pub trait CommitWriter {
    fn accept_commit(
        &mut self,
        domain: &DomainRef,
        scope: &Scope,
        intent: &Intent,
        target: &str,
    ) -> Result<String, AuthorityError>;
}

pub trait ReceiptWriter {
    fn append_receipt(&mut self, receipt: &Receipt) -> Result<(), AuthorityError>;
}

pub trait IdentityBinder {
    fn validate_actor(&self, actor: &ActorRef) -> Result<(), AuthorityError>;
}

// Phase 3.3 — explicit authority to mint receipt identities
pub trait ReceiptIdMint {
    fn mint_receipt_id(&mut self) -> ReceiptRef;
}

// Phase 3.7 — parent lookup for lineage validation
pub trait ReceiptParentReader {
    fn get_receipt(&self, id: &ReceiptRef) -> Result<Option<Receipt>, AuthorityError>;
}

// -----------------------------------------------------------------------------
// Authority Kernel
// -----------------------------------------------------------------------------

pub struct AuthorityKernelConfig {
    pub enforce_non_bypass: bool,

    // Phase 3.7 — bounded lineage walk (no unbounded recursion)
    pub max_parent_hops: usize,
}

pub struct AuthorityKernel<R, W, M> {
    cfg: AuthorityKernelConfig,
    pub reader: R,
    pub writer: W,
    pub minter: M,
}

impl<R, W, M> AuthorityKernel<R, W, M>
where
    R: FreezeStateReader + IdentityBinder + ReceiptParentReader,
    W: FreezeStateWriter + CommitWriter + ReceiptWriter,
    M: ReceiptIdMint,
{
    pub fn new(cfg: AuthorityKernelConfig, reader: R, writer: W, minter: M) -> Self {
        Self { cfg, reader, writer, minter }
    }

    pub fn apply(&mut self, req: AuthorityRequest) -> AuthorityOutcome {
        match req {
            AuthorityRequest::Freeze { actor, intent, policy, provenance, parent } => {
                self.apply_freeze(actor, intent, policy, provenance, parent)
            }
            AuthorityRequest::Commit { actor, intent, policy, provenance, parent } => {
                self.apply_commit(actor, intent, policy, provenance, parent)
            }
        }
    }

    fn apply_freeze(
        &mut self,
        actor: ActorRef,
        intent: FreezeIntent,
        policy: PolicyRef,
        provenance: ProvenanceRef,
        parent: Option<ReceiptRef>,
    ) -> AuthorityOutcome {
        if let Err(e) = self.reader.validate_actor(&actor) {
            return self.emit_error(actor, intent.domain, intent.scope, parent, e, policy, provenance);
        }

        if intent.domain.id.is_empty() || intent.scope.key.is_empty() {
            return self.emit_error(
                actor,
                intent.domain,
                intent.scope,
                parent,
                AuthorityError::InvalidScope,
                policy,
                provenance,
            );
        }

        if let Err(e) = self.validate_parent(&intent.domain, &parent) {
            return self.emit_error(actor, intent.domain, intent.scope, parent, e, policy, provenance);
        }

        match self.writer.apply_freeze_op(
            &intent.domain,
            &intent.scope,
            intent.op,
            intent.ttl_hint_seconds,
        ) {
            Ok(freeze_ref) => {
                let rid = self.minter.mint_receipt_id();
                let receipt = ReceiptMint::allowed(
                    rid.clone(),
                    actor.clone(),
                    intent.domain.clone(),
                    intent.scope.clone(),
                    parent,
                    policy,
                    provenance,
                );
                let _ = self.writer.append_receipt(&receipt);

                AuthorityOutcome::Allowed {
                    receipt: rid,
                    sealed: SealedOutcomeData::FreezeStateRef(freeze_ref),
                }
            }
            Err(e) => self.emit_error(actor, intent.domain, intent.scope, parent, e, policy, provenance),
        }
    }

    fn apply_commit(
        &mut self,
        actor: ActorRef,
        intent: CommitIntent,
        policy: PolicyRef,
        provenance: ProvenanceRef,
        parent: Option<ReceiptRef>,
    ) -> AuthorityOutcome {
        if let Err(e) = self.reader.validate_actor(&actor) {
            return self.emit_error(actor, intent.domain, intent.scope, parent, e, policy, provenance);
        }

        if intent.domain.id.is_empty() || intent.scope.key.is_empty() {
            return self.emit_error(
                actor,
                intent.domain,
                intent.scope,
                parent,
                AuthorityError::InvalidScope,
                policy,
                provenance,
            );
        }

        if let Err(e) = self.validate_parent(&intent.domain, &parent) {
            return self.emit_error(actor, intent.domain, intent.scope, parent, e, policy, provenance);
        }

        match self.reader.is_frozen(&intent.domain, &intent.scope) {
            Ok(true) => {
                let reason = DenyReason::FreezeActive { scope: intent.scope.key.clone() };

                let rid = self.minter.mint_receipt_id();
                let receipt = ReceiptMint::denied(
                    rid.clone(),
                    actor.clone(),
                    intent.domain.clone(),
                    intent.scope.clone(),
                    parent,
                    &reason,
                    policy,
                    provenance,
                );
                let _ = self.writer.append_receipt(&receipt);

                AuthorityOutcome::Denied { receipt: rid, reason }
            }

            Ok(false) => match self.writer.accept_commit(
                &intent.domain,
                &intent.scope,
                &intent.intent,
                &intent.target,
            ) {
                Ok(commit_ref) => {
                    let rid = self.minter.mint_receipt_id();
                    let receipt = ReceiptMint::allowed(
                        rid.clone(),
                        actor.clone(),
                        intent.domain.clone(),
                        intent.scope.clone(),
                        parent,
                        policy,
                        provenance,
                    );
                    let _ = self.writer.append_receipt(&receipt);

                    AuthorityOutcome::Allowed {
                        receipt: rid,
                        sealed: SealedOutcomeData::CommitRef(commit_ref),
                    }
                }
                Err(e) => self.emit_error(actor, intent.domain, intent.scope, parent, e, policy, provenance),
            },

            Err(e) => self.emit_error(actor, intent.domain, intent.scope, parent, e, policy, provenance),
        }
    }

    // -------------------------------------------------------------------------
    // Phase 3.7 — parent validation (bounded)
    // -------------------------------------------------------------------------
    fn validate_parent(
        &self,
        current_domain: &DomainRef,
        parent: &Option<ReceiptRef>,
    ) -> Result<(), AuthorityError> {
        let Some(mut cursor) = parent.clone() else {
            return Ok(());
        };

        // hop-bounded walk; detect cycles by repeats
        let mut seen: std::collections::HashSet<ReceiptRef> = std::collections::HashSet::new();

        for _ in 0..self.cfg.max_parent_hops {
            if !seen.insert(cursor.clone()) {
                return Err(AuthorityError::ParentCycleDetected);
            }

            let Some(r) = self.reader.get_receipt(&cursor)? else {
                return Err(AuthorityError::ParentNotFound);
            };

            if r.domain != *current_domain {
                return Err(AuthorityError::ParentDomainMismatch);
            }

            match r.parent.clone() {
                Some(next) => cursor = next,
                None => return Ok(()),
            }
        }

        Err(AuthorityError::ParentCycleDetected)
    }

    // -------------------------------------------------------------------------
    // Receipt helpers
    // -------------------------------------------------------------------------
    fn emit_error(
        &mut self,
        actor: ActorRef,
        domain: DomainRef,
        scope: Scope,
        parent: Option<ReceiptRef>,
        err: AuthorityError,
        policy: PolicyRef,
        provenance: ProvenanceRef,
    ) -> AuthorityOutcome {
        let rid = self.minter.mint_receipt_id();
        let receipt = ReceiptMint::error(
            rid.clone(),
            actor,
            domain,
            scope,
            parent,
            &err,
            policy,
            provenance,
        );
        let _ = self.writer.append_receipt(&receipt);
        AuthorityOutcome::Error(err)
    }
}
