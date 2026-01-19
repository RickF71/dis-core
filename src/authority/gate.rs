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
    fn apply_freeze_op(&mut self, domain: &DomainRef, scope: &Scope, op: FreezeOp, ttl_hint_seconds: Option<u64>)
        -> Result<String, AuthorityError>; // returns FreezeStateRef
}

pub trait CommitWriter {
    fn accept_commit(&mut self, domain: &DomainRef, scope: &Scope, intent: &Intent, target: &str)
        -> Result<String, AuthorityError>; // returns CommitRef
}

pub trait ReceiptWriter {
    fn append_receipt(&mut self, receipt: &Receipt) -> Result<(), AuthorityError>;
}

pub trait IdentityBinder {
    // Validates that actor exists in Nullus and is eligible to act as Corporeal.
    fn validate_actor(&self, actor: &ActorRef) -> Result<(), AuthorityError>;
}

pub struct AuthorityKernelConfig {
    // Placeholder: add toggles only if strictly needed.
    pub enforce_non_bypass: bool,
}

pub struct AuthorityKernel<R, W> {
    cfg: AuthorityKernelConfig,
    // Readers/writers are injected to keep authority pure and bounded.
    // Bundle R/W generics to avoid leaking runtime types here.
    pub reader: R,
    pub writer: W,
}

impl<R, W> AuthorityKernel<R, W>
where
    R: FreezeStateReader + IdentityBinder,
    W: FreezeStateWriter + CommitWriter + ReceiptWriter,
{
    pub fn new(cfg: AuthorityKernelConfig, reader: R, writer: W) -> Self {
        Self { cfg, reader, writer }
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
            let receipt = ReceiptMint::error(actor.clone(), intent.domain.clone(), intent.scope.clone(), &e, policy, provenance);
            let _ = self.writer.append_receipt(&receipt);
            return AuthorityOutcome::Error(e);
        }

        // Aether ↔ Lima admissibility: handled structurally by types + canonicalization outside.
        // (No Lima fields are consulted for semantics.)

        // Terra ↔ Numen validity: scope/domain basic checks
        if intent.scope.key.is_empty() || intent.domain.id.is_empty() {
            let e = AuthorityError::InvalidScope;
            let receipt = ReceiptMint::error(actor.clone(), intent.domain.clone(), intent.scope.clone(), &e, policy, provenance);
            let _ = self.writer.append_receipt(&receipt);
            return AuthorityOutcome::Error(e);
        }

        // Apply the freeze op (authoritative)
        let apply_res = self.writer.apply_freeze_op(&intent.domain, &intent.scope, intent.op.clone(), intent.ttl_hint_seconds);

        match apply_res {
            Ok(freeze_ref) => {
                let receipt = ReceiptMint::allowed(actor.clone(), intent.domain.clone(), intent.scope.clone(), policy, provenance);
                let _ = self.writer.append_receipt(&receipt);
                AuthorityOutcome::Allowed {
                    receipt: receipt.id.clone(),
                    sealed: SealedOutcomeData::FreezeStateRef(freeze_ref),
                }
            }
            Err(e) => {
                let receipt = ReceiptMint::error(actor.clone(), intent.domain.clone(), intent.scope.clone(), &e, policy, provenance);
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
            let receipt = ReceiptMint::error(actor.clone(), intent.domain.clone(), intent.scope.clone(), &e, policy, provenance);
            let _ = self.writer.append_receipt(&receipt);
            return AuthorityOutcome::Error(e);
        }

        // Freeze gate (minimal authority surface)
        match self.reader.is_frozen(&intent.domain, &intent.scope) {
            Ok(true) => {
                let reason = DenyReason::FreezeActive { scope: intent.scope.key.clone() };
                let receipt = ReceiptMint::denied(actor.clone(), intent.domain.clone(), intent.scope.clone(), &reason, policy, provenance);
                let _ = self.writer.append_receipt(&receipt);
                AuthorityOutcome::Denied { receipt: receipt.id.clone(), reason }
            }
            Ok(false) => {
                // commit acceptance
                match self.writer.accept_commit(&intent.domain, &intent.scope, &intent.intent, &intent.target) {
                    Ok(commit_ref) => {
                        let receipt = ReceiptMint::allowed(actor.clone(), intent.domain.clone(), intent.scope.clone(), policy, provenance);
                        let _ = self.writer.append_receipt(&receipt);
                        AuthorityOutcome::Allowed {
                            receipt: receipt.id.clone(),
                            sealed: SealedOutcomeData::CommitRef(commit_ref),
                        }
                    }
                    Err(e) => {
                        let receipt = ReceiptMint::error(actor.clone(), intent.domain.clone(), intent.scope.clone(), &e, policy, provenance);
                        let _ = self.writer.append_receipt(&receipt);
                        AuthorityOutcome::Error(e)
                    }
                }
            }
            Err(e) => {
                let receipt = ReceiptMint::error(actor.clone(), intent.domain.clone(), intent.scope.clone(), &e, policy, provenance);
                let _ = self.writer.append_receipt(&receipt);
                AuthorityOutcome::Error(e)
            }
        }
    }
}

