// ============================================================
// FILE: src/authority/seal.rs
// ============================================================
//
// Phase 3.2 canonical form:
// - NO receipt ID minting
// - NO globals
// - NO authority
// - Assembly only
//

use super::types::{
    Receipt,
    ReceiptOutcome,
    ReceiptKind,
    ReceiptRef,
    ActorRef,
    DomainRef,
    Scope,
    PolicyRef,
    ProvenanceRef,
};

// Capability token — cannot be constructed outside authority module
pub(crate) struct Sealed;

// Assembly-only function — receipt ID MUST be supplied by authority
pub(crate) fn seal_receipt(
    _sealed: Sealed,
    id: ReceiptRef,
    actor: ActorRef,
    domain: DomainRef,
    scope: Scope,
    outcome: ReceiptOutcome,
    policy: PolicyRef,
    provenance: ProvenanceRef,
) -> Receipt {
    Receipt {
        id,
        kind: ReceiptKind::CiCallV1,
        actor,
        domain,
        scope,
        outcome,
        policy,
        provenance,
        signature: None,
    }
}
