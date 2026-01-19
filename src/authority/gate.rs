// ============================================================
// FILE: src/authority/gate.rs
// ============================================================

use super::types::*;
use super::errors::*;
use super::receipt::ReceiptMint;

// Injected traits (implemented elsewhere) — still quarantined.
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
    ) -> Result<String, AuthorityError>; // returns FreezeStateRef
}

pub trait CommitWriter {
    fn accept_commit(
        &mut self,
        domain: &DomainRef,
        scope: &Scope,
        intent: &Intent,
        target: &str,
    ) -> Result<String, AuthorityError>; // returns CommitRef
}

pub trait ReceiptWriter {
    fn append_receipt(&mut self, receipt: &Receipt) -> Result<(), AuthorityError>;
}

pub trait IdentityBinder {
    // Validates that actor exists in Nullus and is eligible to act as Corporeal.
    fn validate_actor(&self, actor: &ActorRef) -> Result<(), AuthorityError>;
}

// Phase 3.2: receipt identity is an explicit authority capability.
pub trait ReceiptIdMint {
    fn mint_receipt_id(&mut self) -> ReceiptRef;
}

pub struct AuthorityKernelConfig {
    // Placeholder: add toggles only if strictly needed.
    pub enforce_non_bypass: bool,
}

pub struct AuthorityKernel<R, W, M> {
    cfg: AuthorityKernelConfig,
    // Readers/writers are injected to keep authority pure and bounded.
    pub reader: R,
    pub writer: W,
    pub minter: M,
}

impl<R, W, M> AuthorityKernel<R, W, M>
where
    R: FreezeStateReader + IdentityBinder,
    W: FreezeStateWriter + CommitWriter + ReceiptWriter,
    M: ReceiptIdMint,
{
    pub fn new(cfg: AuthorityKernelConfig, reader: R, writer: W, minter: M) -> Self {
        Self { cfg, reader, writer, minter }
    }

    // The only public entry: evaluate + commit + receipt.
    pub fn apply(&mut self, req: AuthorityRequest) -> AuthorityOutcome {
        match req {
            AuthorityRequest::Freeze { actor, intent, policy, provenance } => {
                self.apply_freeze(actor, intent, policy, provenance)
            }
            AuthorityRequest::Commit { actor, intent, policy, provenance } => {
                self.apply_commit(actor, intent, policy, provenance)
            }
        }
    }

    fn apply_freeze(
        &mut self,
        actor: ActorRef,
        intent: FreezeIntent,
        policy: PolicyRef,
        provenance: ProvenanceRef,
    ) -> AuthorityOutcome {
        // Nullus ↔ Corporeal closure (validate actor)
        if let Err(e) = self.reader.validate_actor(&actor) {
            let rid = self.minter.mint_receipt_id();
            let receipt = ReceiptMint::error(
                rid.clone(),
                actor.clone(),
                intent.domain.clone(),
                intent.scope.clone(),
                &e,
                policy,
                provenance,
            );
            let _ = self.writer.append_receipt(&receipt);
            return AuthorityOutcome::Error(e);
        }

        // Terra ↔ Numen validity: scope/domain basic checks
        if intent.scope.key.is_empty() || intent.domain.id.is_empty() {
            let e = AuthorityError::InvalidScope;
            let rid = self.minter.mint_receipt_id();
            let receipt = ReceiptMint::error(
                rid.clone(),
                actor.clone(),
                intent.domain.clone(),
                intent.scope.clone(),
                &e,
                policy,
                provenance,
            );
            let _ = self.writer.append_receipt(&receipt);
            return AuthorityOutcome::Error(e);
        }

        // Apply the freeze op (authoritative)
        let apply_res = self.writer.apply_freeze_op(
            &intent.domain,
            &intent.scope,
            intent.op.clone(),
            intent.ttl_hint_seconds,
        );

        match apply_res {
            Ok(freeze_ref) => {
                let rid = self.minter.mint_receipt_id();
                let receipt = ReceiptMint::allowed(
                    rid.clone(),
                    actor.clone(),
                    intent.domain.clone(),
                    intent.scope.clone(),
                    policy,
                    provenance,
                );
                let _ = self.writer.append_receipt(&receipt);
                AuthorityOutcome::Allowed {
                    receipt: rid,
                    sealed: SealedOutcomeData::FreezeStateRef(freeze_ref),
                }
            }
            Err(e) => {
                let rid = self.minter.mint_receipt_id();
                let receipt = ReceiptMint::error(
                    rid.clone(),
                    actor.clone(),
                    intent.domain.clone(),
                    intent.scope.clone(),
                    &e,
                    policy,
                    provenance,
                );
                let _ = self.writer.append_receipt(&receipt);
                AuthorityOutcome::Error(e)
            }
        }
    }

    fn apply_commit(
        &mut self,
        actor: ActorRef,
        intent: CommitIntent,
        policy: PolicyRef,
        provenance: ProvenanceRef,
    ) -> AuthorityOutcome {
        // Nullus ↔ Corporeal closure
        if let Err(e) = self.reader.validate_actor(&actor) {
            let rid = self.minter.mint_receipt_id();
            let receipt = ReceiptMint::error(
                rid.clone(),
                actor.clone(),
                intent.domain.clone(),
                intent.scope.clone(),
                &e,
                policy,
                provenance,
            );
            let _ = self.writer.append_receipt(&receipt);
            return AuthorityOutcome::Error(e);
        }

        // Freeze gate (minimal authority surface)
        match self.reader.is_frozen(&intent.domain, &intent.scope) {
            Ok(true) => {
                let reason = DenyReason::FreezeActive { scope: intent.scope.key.clone() };

                let rid = self.minter.mint_receipt_id();
                let receipt = ReceiptMint::denied(
                    rid.clone(),
                    actor.clone(),
                    intent.domain.clone(),
                    intent.scope.clone(),
                    &reason,
                    policy,
                    provenance,
                );
                let _ = self.writer.append_receipt(&receipt);

                AuthorityOutcome::Denied { receipt: rid, reason }
            }
            Ok(false) => {
                // commit acceptance
                match self.writer.accept_commit(&intent.domain, &intent.scope, &intent.intent, &intent.target) {
                    Ok(commit_ref) => {
                        let rid = self.minter.mint_receipt_id();
                        let receipt = ReceiptMint::allowed(
                            rid.clone(),
                            actor.clone(),
                            intent.domain.clone(),
                            intent.scope.clone(),
                            policy,
                            provenance,
                        );
                        let _ = self.writer.append_receipt(&receipt);

                        AuthorityOutcome::Allowed {
                            receipt: rid,
                            sealed: SealedOutcomeData::CommitRef(commit_ref),
                        }
                    }
                    Err(e) => {
                        let rid = self.minter.mint_receipt_id();
                        let receipt = ReceiptMint::error(
                            rid.clone(),
                            actor.clone(),
                            intent.domain.clone(),
                            intent.scope.clone(),
                            &e,
                            policy,
                            provenance,
                        );
                        let _ = self.writer.append_receipt(&receipt);

                        AuthorityOutcome::Error(e)
                    }
                }
            }
            Err(e) => {
                let rid = self.minter.mint_receipt_id();
                let receipt = ReceiptMint::error(
                    rid.clone(),
                    actor.clone(),
                    intent.domain.clone(),
                    intent.scope.clone(),
                    &e,
                    policy,
                    provenance,
                );
                let _ = self.writer.append_receipt(&receipt);

                AuthorityOutcome::Error(e)
            }
        }
    }
}
