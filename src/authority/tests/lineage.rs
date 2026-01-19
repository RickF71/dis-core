use std::cell::RefCell;
use std::rc::Rc;

use crate::authority::*;
use super::fakes::{FakeAuthorityBackend, SharedBackend};

#[test]
fn receipt_preserves_parent_lineage() {
    let backend: SharedBackend = Rc::new(RefCell::new(
        FakeAuthorityBackend::default().with_actor("alice")
    ));

    let mut kernel = AuthorityKernel::new(
        AuthorityKernelConfig {
            enforce_non_bypass: true,
            max_parent_hops: 64,
        },
        backend.clone(), // reader
        backend.clone(), // writer
        super::fakes::TestReceiptMinter::default(),
    );

    let actor = ActorRef { id: "alice".into() };
    let domain = DomainRef { id: "domain.test".into() };
    let scope = Scope { key: "root".into() };

    // ---------------------------------------------------------
    // 1) First commit — produces a REAL parent receipt
    // ---------------------------------------------------------
    let out1 = kernel.apply(AuthorityRequest::Commit {
        actor: actor.clone(),
        intent: CommitIntent {
            domain: domain.clone(),
            scope: scope.clone(),
            intent: Intent { canonical_hash: [1u8; 32] },
            target: "artifact-parent".into(),
        },
        policy: PolicyRef { id: "policy.ok".into() },
        provenance: ProvenanceRef { id: "prov.ok".into() },
        parent: None,
    });

    let parent_receipt = match out1 {
        AuthorityOutcome::Allowed { receipt, .. } => receipt,
        other => panic!("expected Allowed for parent commit, got {:?}", other),
    };

    // ---------------------------------------------------------
    // 2) Second commit — references the real parent receipt
    // ---------------------------------------------------------
    let out2 = kernel.apply(AuthorityRequest::Commit {
        actor,
        intent: CommitIntent {
            domain,
            scope,
            intent: Intent { canonical_hash: [2u8; 32] },
            target: "artifact-child".into(),
        },
        policy: PolicyRef { id: "policy.ok".into() },
        provenance: ProvenanceRef { id: "prov.ok".into() },
        parent: Some(parent_receipt.clone()),
    });

    let child_receipt_id = match out2 {
        AuthorityOutcome::Allowed { receipt, .. } => receipt,
        other => panic!("expected Allowed for child commit, got {:?}", other),
    };

    // ---------------------------------------------------------
    // 3) Assert lineage is preserved in the stored receipt
    // ---------------------------------------------------------
    let backend_ref = backend.borrow();
    let child_receipt = backend_ref
        .receipts
        .iter()
        .find(|r| r.id == child_receipt_id)
        .expect("child receipt missing");

    assert_eq!(
        child_receipt.parent,
        Some(parent_receipt),
        "child receipt must preserve parent lineage"
    );
}
