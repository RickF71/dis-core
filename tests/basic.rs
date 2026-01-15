use triad::spine::{
    lima::{Transition, LimaContext, LimaEvaluator, Decision},
    terra::{TerraNodeId},
    numen::MeaningId,
    ActorId,
    DomainId,
    SeatId,
};

use std::collections::HashMap;

// --- Minimal TerraView ---
struct TestTerra;

impl triad::spine::lima::TerraView for TestTerra {
    fn has_edge(&self, _from: &TerraNodeId, _to: &TerraNodeId) -> bool {
        true
    }
}

// --- Minimal NumenView ---
struct TestNumen;

impl triad::spine::lima::NumenView for TestNumen {
    fn meaning_applies(&self, _t: &Transition) -> bool {
        true
    }
}

#[test]
fn lima_allows_simple_transition() {
    let domain = DomainId(1);
    let actor  = ActorId(100);
    let seat   = SeatId(10);


    let from = TerraNodeId("A".into());
    let to   = TerraNodeId("B".into());

    let meaning = MeaningId("test.move".into());

    // --- Aether bindings ---
    let mut aether = triad::spine::lima::AetherBindings::new();
    aether.bind(seat.clone(), actor.clone());

    // --- Context ---
    let ctx = LimaContext {
        domain: domain.clone(),
        aether: &aether,
        terra: &TestTerra,
        numen: &TestNumen,
        turn: triad::spine::lima::TurnState::none(),
        freeze: triad::spine::lima::FreezeState::none(),
        domain_rules: None,
    };

    let t = Transition::new(
        domain,
        actor,
        seat,
        from,
        to,
        meaning,
        triad::spine::lima::TickId(1),
    );

    let decision = LimaEvaluator::evaluate(&t, &ctx);

    assert!(matches!(decision, Decision::Allow));
}
