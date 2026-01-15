//
// Lima — Transition Receipt
//
// Immutable record that a Lima evaluation occurred.
// Receipts prove authority was exercised; they do not grant authority.
//

use crate::spine::{
    SpineLayer,
    terra::TerraNodeId,
    numen::MeaningId,
    ActorId,
    DomainId,
    SeatId,
};

use super::{
    Decision,
    TickId,
};

/// Receipt emitted by Lima after evaluating a transition.
#[derive(Debug, Clone)]
pub struct TransitionReceipt {
    pub domain: DomainId,
    pub actor: ActorId,
    pub seat: SeatId,

    pub from: TerraNodeId,
    pub to: TerraNodeId,
    pub meaning: MeaningId,

    pub tick: TickId,
    pub layer: SpineLayer,

    pub decision: Decision,
}

impl TransitionReceipt {
    pub fn from_decision(
        t: &super::Transition,
        decision: Decision,
    ) -> Self {
        Self {
            domain: t.domain.clone(),
            actor: t.actor.clone(),
            seat: t.seat.clone(),
            from: t.from.clone(),
            to: t.to.clone(),
            meaning: t.meaning.clone(),
            tick: t.tick,
            layer: SpineLayer::Lima,
            decision,
        }
    }
}
