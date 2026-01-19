use std::cell::RefCell;
use std::rc::Rc;

use crate::authority::*;
use super::fakes::{FakeAuthorityBackend, SharedBackend};

#[test]
fn parent_not_found_errors() {
    let backend: SharedBackend = Rc::new(RefCell::new(
        FakeAuthorityBackend::default().with_actor("alice")
    ));

    let mut kernel = AuthorityKernel::new(
        AuthorityKernelConfig { enforce_non_bypass: true, max_parent_hops: 64 },
        backend.clone(),
        backend.clone(),
        super::fakes::TestReceiptMinter::default(),
    );

    let missing_parent = ReceiptRef { id: "rcpt-missing".into() };

    let out = kernel.apply(AuthorityRequest::Commit {
        actor: ActorRef { id: "alice".into() },
        intent: CommitIntent {
            domain: DomainRef { id: "domain.test".into() },
            scope: Scope { key: "root".into() },
            intent: Intent { canonical_hash: [7u8; 32] },
            target: "artifact".into(),
        },
        policy: PolicyRef { id: "policy.ok".into() },
        provenance: ProvenanceRef { id: "prov.ok".into() },
        parent: Some(missing_parent.clone()),
    });

    assert!(matches!(out, AuthorityOutcome::Error(AuthorityError::ParentNotFound)));

    // receipt exists and carries attempted parent
    let b = backend.borrow();
    let last = b.receipts.last().expect("receipt missing");
    assert_eq!(last.parent, Some(missing_parent));
}

#[test]
fn parent_domain_mismatch_errors() {
    let backend: SharedBackend = Rc::new(RefCell::new(
        FakeAuthorityBackend::default().with_actor("alice")
    ));

    // seed a parent receipt in a different domain
    backend.borrow_mut().receipts.push(Receipt {
        id: ReceiptRef { id: "rcpt-foreign".into() },
        kind: ReceiptKind::CiCallV1,
        actor: ActorRef { id: "alice".into() },
        domain: DomainRef { id: "domain.other".into() },
        scope: Scope { key: "root".into() },
        parent: None,
        outcome: ReceiptOutcome::Allowed,
        policy: PolicyRef { id: "policy.ok".into() },
        provenance: ProvenanceRef { id: "prov.ok".into() },
        signature: None,
    });

    let mut kernel = AuthorityKernel::new(
        AuthorityKernelConfig { enforce_non_bypass: true, max_parent_hops: 64 },
        backend.clone(),
        backend.clone(),
        super::fakes::TestReceiptMinter::default(),
    );

    let out = kernel.apply(AuthorityRequest::Commit {
        actor: ActorRef { id: "alice".into() },
        intent: CommitIntent {
            domain: DomainRef { id: "domain.test".into() },
            scope: Scope { key: "root".into() },
            intent: Intent { canonical_hash: [8u8; 32] },
            target: "artifact".into(),
        },
        policy: PolicyRef { id: "policy.ok".into() },
        provenance: ProvenanceRef { id: "prov.ok".into() },
        parent: Some(ReceiptRef { id: "rcpt-foreign".into() }),
    });

    assert!(matches!(out, AuthorityOutcome::Error(AuthorityError::ParentDomainMismatch)));
}

#[test]
fn parent_cycle_errors() {
    let backend: SharedBackend = Rc::new(RefCell::new(
        FakeAuthorityBackend::default().with_actor("alice")
    ));

    // seed a cycle: rcpt-a -> rcpt-a
    backend.borrow_mut().receipts.push(Receipt {
        id: ReceiptRef { id: "rcpt-a".into() },
        kind: ReceiptKind::CiCallV1,
        actor: ActorRef { id: "alice".into() },
        domain: DomainRef { id: "domain.test".into() },
        scope: Scope { key: "root".into() },
        parent: Some(ReceiptRef { id: "rcpt-a".into() }),
        outcome: ReceiptOutcome::Allowed,
        policy: PolicyRef { id: "policy.ok".into() },
        provenance: ProvenanceRef { id: "prov.ok".into() },
        signature: None,
    });

    let mut kernel = AuthorityKernel::new(
        AuthorityKernelConfig { enforce_non_bypass: true, max_parent_hops: 64 },
        backend.clone(),
        backend.clone(),
        super::fakes::TestReceiptMinter::default(),
    );

    let out = kernel.apply(AuthorityRequest::Commit {
        actor: ActorRef { id: "alice".into() },
        intent: CommitIntent {
            domain: DomainRef { id: "domain.test".into() },
            scope: Scope { key: "root".into() },
            intent: Intent { canonical_hash: [9u8; 32] },
            target: "artifact".into(),
        },
        policy: PolicyRef { id: "policy.ok".into() },
        provenance: ProvenanceRef { id: "prov.ok".into() },
        parent: Some(ReceiptRef { id: "rcpt-a".into() }),
    });

    assert!(matches!(out, AuthorityOutcome::Error(AuthorityError::ParentCycleDetected)));
}
