// ============================================================
// FILE: src/authority/tests/envelope.rs
// ============================================================

use crate::authority::*;

#[test]
fn envelope_builder_wraps_core_and_keeps_extras() {
    let core = Receipt {
        id: ReceiptRef { id: "rcpt-1".into() },
        kind: ReceiptKind::CiCallV1,
        actor: ActorRef { id: "alice".into() },
        domain: DomainRef { id: "domain.test".into() },
        scope: Scope { key: "root".into() },
        parent: None,
        outcome: ReceiptOutcome::Allowed,
        policy: PolicyRef { id: "policy.ok".into() },
        provenance: ProvenanceRef { id: "prov.ok".into() },
        signature: None,
    };

    let env = ReceiptEnvelopeBuilder::new(core.clone())
        .add_seal(ReceiptSeal {
            scheme: "ed25519".into(),
            key_id: "key-1".into(),
            sig: vec![1, 2, 3],
        })
        .add_witness(WitnessRef { id: "witness-1".into() })
        .build();

    assert_eq!(env.core, core);
    assert_eq!(env.seals.len(), 1);
    assert_eq!(env.witnesses.len(), 1);
}

#[test]
fn signable_v2_includes_parent_when_present() {
    let parent = ReceiptRef { id: "rcpt-parent".into() };

    let r = Receipt {
        id: ReceiptRef { id: "rcpt-child".into() },
        kind: ReceiptKind::CiCallV1,
        actor: ActorRef { id: "alice".into() },
        domain: DomainRef { id: "domain.test".into() },
        scope: Scope { key: "root".into() },
        parent: Some(parent),
        outcome: ReceiptOutcome::Denied { code: "deny:freeze:root".into() },
        policy: PolicyRef { id: "policy.ok".into() },
        provenance: ProvenanceRef { id: "prov.ok".into() },
        signature: None,
    };

    let s = r.signable_string_v2();
    assert!(s.contains("|parent=rcpt-parent|"), "v2 must bind parent");
    assert!(s.contains("receipt.signable.v2|"), "must be v2 marker");
}
