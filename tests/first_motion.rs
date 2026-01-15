use triad::spine::{
    lima::{
        Transition,
        LimaContext,
        LimaEvaluator,
        Decision,
        TickId,
        AetherBindings,
        TurnState,
        FreezeState,
        TerraView,
        NumenView,
    },
    terra::TerraNodeId,
    numen::MeaningId,
    ActorId,
    DomainId,
    SeatId,
};

use std::collections::HashSet;

// ------------------------------------------------------
// Minimal Terra: "everything is adjacent"
// ------------------------------------------------------

struct TestTerra;

impl TerraView for TestTerra {
    fn has_edge(&self, _from: &TerraNodeId, _to: &TerraNodeId) -> bool {
        true
    }
}

// ------------------------------------------------------
// Minimal Numen: "meaning always applies"
// ------------------------------------------------------

struct TestNumen;

impl NumenView for TestNumen {
    fn meaning_applies(&self, _t: &Transition) -> bool {
        true
    }

    fn contracts_for_meaning(&self, _meaning: &MeaningId) -> HashSet<triad::spine::numen::ContractId> {
        HashSet::new()
    }
}

// ------------------------------------------------------
// The test: first real motion through the spine
// ------------------------------------------------------

#[test]
fn first_lima_motion_is_allowed() {
    // --- Identity ---
    let domain = DomainId(1);
    let actor  = ActorId(42);
    let seat   = SeatId(7);


    // --- Structure ---
    let from = TerraNodeId("A".into());
    let to   = TerraNodeId("B".into());

    // --- Meaning ---
    let meaning = MeaningId("move.test".into());

    // --- Time ---
    let tick = TickId(1);

    // --- Aether binding ---
    let mut aether = AetherBindings::new();
    aether.bind(seat.clone(), actor.clone());

    // --- Context ---
    let terra = TestTerra;
    let numen = TestNumen;

    let ctx = LimaContext {
        domain: domain.clone(),
        aether: &aether,
        terra: &terra,
        numen: &numen,
        turn: TurnState::none(),
        freeze: FreezeState::none(),
        domain_rules: None,
    };

    // --- Hypothesis ---
    let transition = Transition::new(
        domain,
        actor,
        seat,
        from,
        to,
        meaning,
        tick,
    );

    // --- Evaluation ---
    let decision = LimaEvaluator::evaluate(&transition, &ctx);

    // --- Assertion ---
    assert!(
        matches!(decision, Decision::Allow),
        "expected transition to be allowed, got {:?}",
        decision
    );
}
