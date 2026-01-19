// ============================================================
// FILE: src/authority/seal.rs
// ============================================================

use std::sync::atomic::{AtomicU64, Ordering};

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

// -----------------------------------------------------------------------------
// Receipt identity — deterministic, monotonic, no clocks, no entropy
// -----------------------------------------------------------------------------
static NEXT_RECEIPT_ID: AtomicU64 = AtomicU64::new(1);

// This is the capability token — MUST exist
pub(crate) struct Sealed;

// This is the function receipt.rs imports — MUST match name exactly
pub(crate) fn seal_receipt(
    _sealed: Sealed,
    actor: ActorRef,
    domain: DomainRef,
    scope: Scope,
    outcome: ReceiptOutcome,
    policy: PolicyRef,
    provenance: ProvenanceRef,
) -> Receipt {
    let n = NEXT_RECEIPT_ID.fetch_add(1, Ordering::Relaxed);

    let id = ReceiptRef {
        id: format!("rcpt-{}", n),
    };

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
