// ============================================================
// FILE: src/authority/receipt.rs
// ============================================================
//
// Phase 3.2 canonical form:
// - ReceiptMint does NOT mint IDs
// - ReceiptRef is supplied by authority (gate.rs)
// - This module formats receipts only
//

use super::seal::{Sealed, seal_receipt};
use super::types::*;
use super::errors::*;

pub(crate) struct ReceiptMint;

impl ReceiptMint {
    pub(crate) fn allowed(
        id: ReceiptRef,
        actor: ActorRef,
        domain: DomainRef,
        scope: Scope,
        policy: PolicyRef,
        provenance: ProvenanceRef,
    ) -> Receipt {
        seal_receipt(
            Sealed,
            id,
            actor,
            domain,
            scope,
            ReceiptOutcome::Allowed,
            policy,
            provenance,
        )
    }

    pub(crate) fn denied(
        id: ReceiptRef,
        actor: ActorRef,
        domain: DomainRef,
        scope: Scope,
        reason: &DenyReason,
        policy: PolicyRef,
        provenance: ProvenanceRef,
    ) -> Receipt {
        let code = match reason {
            DenyReason::FreezeActive { scope } => format!("deny:freeze:{}", scope),
            DenyReason::PolicyDenied { .. } => "deny:policy".to_string(),
            DenyReason::ActorNotAuthorized => "deny:actor".to_string(),
            DenyReason::ScopeNotAdopted => "deny:scope".to_string(),
            DenyReason::Other { code } => code.clone(),
        };

        seal_receipt(
            Sealed,
            id,
            actor,
            domain,
            scope,
            ReceiptOutcome::Denied { code },
            policy,
            provenance,
        )
    }

    pub(crate) fn error(
        id: ReceiptRef,
        actor: ActorRef,
        domain: DomainRef,
        scope: Scope,
        err: &AuthorityError,
        policy: PolicyRef,
        provenance: ProvenanceRef,
    ) -> Receipt {
        let code = match err {
            AuthorityError::MissingIdentityBinding => "err:missing_identity",
            AuthorityError::NonBypassViolation => "err:non_bypass",
            AuthorityError::InvalidScope => "err:invalid_scope",
            AuthorityError::InvalidPolicyRef => "err:invalid_policy_ref",
            AuthorityError::InvalidProvenanceRef => "err:invalid_provenance_ref",
            AuthorityError::KernelMisconfiguration => "err:kernel_misconfig",
            AuthorityError::InternalInvariantFailed(_) => "err:invariant",
        }
        .to_string();

        seal_receipt(
            Sealed,
            id,
            actor,
            domain,
            scope,
            ReceiptOutcome::Error { code },
            policy,
            provenance,
        )
    }
}
