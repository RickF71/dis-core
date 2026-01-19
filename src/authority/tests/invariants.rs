// ============================================================
// FILE: src/authority/tests/invariants.rs
// ============================================================

use std::cell::RefCell;
use std::rc::Rc;

use crate::authority::*;
use super::fakes::{FakeAuthorityBackend, SharedBackend};

fn mk_kernel_with_backend() -> (
    AuthorityKernel<SharedBackend, SharedBackend, super::fakes::TestReceiptMinter>,
    SharedBackend,
) {
    let backend: SharedBackend = Rc::new(RefCell::new(
        FakeAuthorityBackend::default().with_actor("alice")
    ));

    let kernel = AuthorityKernel::new(
        AuthorityKernelConfig { enforce_non_bypass: true, max_parent_hops: 64 },
        backend.clone(), // reader
        backend.clone(), // writer
        super::fakes::TestReceiptMinter::default(),
    );

    (kernel, backend)
}

fn receipt_count(backend: &SharedBackend) -> usize {
    backend.borrow().receipts.len()
}

fn assert_receipt_appended_exactly_one(
    backend: &SharedBackend,
    before: usize,
    after_outcome: &AuthorityOutcome,
) {
    let after = receipt_count(backend);
    assert_eq!(after, before + 1, "apply() must append exactly one receipt");

    // Ensure the outcome's receipt ref exists in the backend receipts
    let receipt_ref = match after_outcome {
        AuthorityOutcome::Allowed { receipt, .. } => receipt,
        AuthorityOutcome::Denied { receipt, .. } => receipt,
        AuthorityOutcome::Error(_) => return, // your AuthorityOutcome::Error currently doesn't carry ReceiptRef
    };

    let b = backend.borrow();
    let matches: Vec<_> = b.receipts.iter().filter(|r| &r.id == receipt_ref).collect();
    assert_eq!(matches.len(), 1, "outcome receipt ref must exist exactly once");
}

#[test]
fn every_apply_emits_exactly_one_receipt() {
    let (mut kernel, backend) = mk_kernel_with_backend();

    let actor = ActorRef { id: "alice".into() };
    let domain = DomainRef { id: "domain.test".into() };
    let scope = Scope { key: "root".into() };

    // 1) Commit when not frozen => Allowed, +1 receipt
    let before = receipt_count(&backend);
    let out = kernel.apply(AuthorityRequest::Commit {
        actor: actor.clone(),
        intent: CommitIntent {
            domain: domain.clone(),
            scope: scope.clone(),
            intent: Intent { canonical_hash: [1u8; 32] },
            target: "artifact-allowed".into(),
        },
        policy: PolicyRef { id: "policy.ok".into() },
        provenance: ProvenanceRef { id: "prov.ok".into() },
        parent: None,
    });
    assert_receipt_appended_exactly_one(&backend, before, &out);

    // 2) Freeze => Allowed, +1 receipt
    let before = receipt_count(&backend);
    let out = kernel.apply(AuthorityRequest::Freeze {
        actor: actor.clone(),
        intent: FreezeIntent {
            op: FreezeOp::Freeze,
            domain: domain.clone(),
            scope: scope.clone(),
            ttl_hint_seconds: None,
            reason: "maintenance".into(),
        },
        policy: PolicyRef { id: "policy.ok".into() },
        provenance: ProvenanceRef { id: "prov.ok".into() },
        parent: None,
    });
    assert_receipt_appended_exactly_one(&backend, before, &out);

    // 3) Commit when frozen => Denied, +1 receipt
    let before = receipt_count(&backend);
    let out = kernel.apply(AuthorityRequest::Commit {
        actor,
        intent: CommitIntent {
            domain,
            scope,
            intent: Intent { canonical_hash: [2u8; 32] },
            target: "artifact-denied".into(),
        },
        policy: PolicyRef { id: "policy.ok".into() },
        provenance: ProvenanceRef { id: "prov.ok".into() },
        parent: None,
    });
    assert_receipt_appended_exactly_one(&backend, before, &out);
}
