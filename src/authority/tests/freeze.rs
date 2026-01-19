use std::rc::Rc;
use std::cell::RefCell;

use crate::authority::*;
use crate::authority::types::ReceiptOutcome;
use super::fakes::{FakeAuthorityBackend, SharedBackend};

#[test]
fn freeze_denies_commit() {
    let actor = ActorRef { id: "alice".into() };
    let domain = DomainRef { id: "domain.test".into() };
    let scope = Scope { key: "root".into() };

    let backend: SharedBackend = Rc::new(RefCell::new(
        FakeAuthorityBackend::default().with_actor("alice")
    ));

    let mut kernel = AuthorityKernel::new(
        AuthorityKernelConfig { enforce_non_bypass: true },
        backend.clone(), // reader
        backend.clone(), // writer
        super::fakes::TestReceiptMinter::default(),
    );


    // Freeze
    let freeze = AuthorityRequest::Freeze {
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
    };

    let out = kernel.apply(freeze);
    assert!(matches!(out, AuthorityOutcome::Allowed { .. }));

    // Commit should now be denied
    let commit = AuthorityRequest::Commit {
        actor,
        intent: CommitIntent {
            domain,
            scope,
            intent: Intent { canonical_hash: [0u8; 32] },
            target: "artifact-1".into(),
        },
        policy: PolicyRef { id: "policy.ok".into() },
        provenance: ProvenanceRef { id: "prov.ok".into() },
    };

    let out = kernel.apply(commit);

    // 1. Authority outcome
    let receipt_ref = match out {
        AuthorityOutcome::Denied { receipt, .. } => receipt,
        other => panic!("expected Denied outcome, got {:?}", other),
    };

    // 2. Receipt must exist
    let backend_ref = backend.borrow();
    let receipt = backend_ref
        .receipts
        .iter()
        .find(|r| r.id == receipt_ref)
        .expect("receipt not written");

    // 3. Receipt outcome must encode freeze denial
    match &receipt.outcome {
        ReceiptOutcome::Denied { code } => {
            assert_eq!(code, "deny:freeze:root");
        }
        other => panic!("expected Denied receipt outcome, got {:?}", other),
    }

    // 4. Actor / scope / domain correctness
    assert_eq!(receipt.actor.id, "alice");
    assert_eq!(receipt.domain.id, "domain.test");
    assert_eq!(receipt.scope.key, "root");

    // 5. Policy & provenance preserved
    assert_eq!(receipt.policy.id, "policy.ok");
    assert_eq!(receipt.provenance.id, "prov.ok");

}


