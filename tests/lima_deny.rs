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
        DenyReason,
        FreezeScopeId,
    },
    terra::TerraNodeId,
    numen::{MeaningId, ContractId},
    ActorId,
    DomainId,
    SeatId,
};

use std::collections::HashSet;

// ------------------------------------------------------
// Minimal Terra
// ------------------------------------------------------

struct TestTerra;

impl TerraView for TestTerra {
    fn has_edge(&self, _from: &TerraNodeId, _to: &TerraNodeId) -> bool {
        true
    }
}

// ------------------------------------------------------
// Minimal Numen
// ------------------------------------------------------

struct TestNumen;

impl NumenView for TestNumen {
    fn meaning_applies(&self, _t: &Transition) -> bool {
        true
    }

    fn contracts_for_meaning(
        &self,
        _meaning: &MeaningId,
    ) -> HashSet<ContractId> {
        HashSet::new()
    }
}

// ------------------------------------------------------
// Tests
// ------------------------------------------------------

#[test]
fn deny_when_seat_not_bound() {
    let domain = DomainId(1);
    let actor  = ActorId(100);
    let seat   = SeatId(10);



    let from = TerraNodeId("A".into());
    let to   = TerraNodeId("B".into());
    let meaning = MeaningId("move.test".into());

    let aether = AetherBindings::new(); // no binding

    let ctx = LimaContext {
        domain: domain.clone(),
        aether: &aether,
        terra: &TestTerra,
        numen: &TestNumen,
        turn: TurnState::none(),
        freeze: FreezeState::none(),
        domain_rules: None,
    };

    let t = Transition::new(
        domain,
        actor,
        seat,
        from,
        to,
        meaning,
        TickId(1),
    );

    let decision = LimaEvaluator::evaluate(&t, &ctx);

    assert!(
        matches!(
            decision,
            Decision::Deny { reason: DenyReason::SeatNotBound { .. } }
        ),
        "expected SeatNotBound, got {:?}",
        decision
    );
}

#[test]
fn deny_when_actor_not_occupant() {
    let domain = DomainId(1);
    let actor_alice  = ActorId(100);
    let actor_bob    = ActorId(101);
    let seat   = SeatId(10);


    let from = TerraNodeId("A".into());
    let to   = TerraNodeId("B".into());
    let meaning = MeaningId("move.test".into());

    let mut aether = AetherBindings::new();
    aether.bind(seat.clone(), actor_alice);

    let ctx = LimaContext {
        domain: domain.clone(),
        aether: &aether,
        terra: &TestTerra,
        numen: &TestNumen,
        turn: TurnState::none(),
        freeze: FreezeState::none(),
        domain_rules: None,
    };

    let t = Transition::new(
        domain,
        actor_bob, // wrong actor
        seat,
        from,
        to,
        meaning,
        TickId(1),
    );

    let decision = LimaEvaluator::evaluate(&t, &ctx);

    assert!(
        matches!(
            decision,
            Decision::Deny { reason: DenyReason::ActorNotOccupant { .. } }
        ),
        "expected ActorNotOccupant, got {:?}",
        decision
    );
}

#[test]
fn deny_when_domain_is_frozen() {
    let domain = DomainId(1);
    let actor  = ActorId(100);
    let seat   = SeatId(10);


    let from = TerraNodeId("A".into());
    let to   = TerraNodeId("B".into());
    let meaning = MeaningId("move.test".into());

    let mut aether = AetherBindings::new();
    aether.bind(seat.clone(), actor.clone());

    let mut freeze = FreezeState::none();
    freeze
        .active_scopes
        .insert(FreezeScopeId("domain".into()));

    let ctx = LimaContext {
        domain: domain.clone(),
        aether: &aether,
        terra: &TestTerra,
        numen: &TestNumen,
        turn: TurnState::none(),
        freeze,
        domain_rules: None,
    };

    let t = Transition::new(
        domain,
        actor,
        seat,
        from,
        to,
        meaning,
        TickId(1),
    );

    let decision = LimaEvaluator::evaluate(&t, &ctx);

    assert!(
        matches!(
            decision,
            Decision::Deny { reason: DenyReason::Frozen { .. } }
        ),
        "expected Frozen deny, got {:?}",
        decision
    );
}
