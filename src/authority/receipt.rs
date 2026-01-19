// ============================================================
// FILE: src/authority/receipt.rs
// ============================================================
//
// Phase 3.7: error code mapping includes lineage validation errors
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
        parent: Option<ReceiptRef>,
        policy: PolicyRef,
        provenance: ProvenanceRef,
    ) -> Receipt {
        seal_receipt(
            Sealed,
            id,
            actor,
            domain,
            scope,
            parent,
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
        parent: Option<ReceiptRef>,
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
            parent,
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
        parent: Option<ReceiptRef>,
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

            // Phase 3.7
            AuthorityError::ParentNotFound => "err:parent_not_found",
            AuthorityError::ParentDomainMismatch => "err:parent_domain_mismatch",
            AuthorityError::ParentCycleDetected => "err:parent_cycle",
        }
        .to_string();

        seal_receipt(
            Sealed,
            id,
            actor,
            domain,
            scope,
            parent,
            ReceiptOutcome::Error { code },
            policy,
            provenance,
        )
    }
}
