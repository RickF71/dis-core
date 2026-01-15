// src/spine/05_lima/transition.rs
use crate::spine::lima::types::*;

//
// A proposed transition (a hypothesis).
// “Suppose this actor, in this seat, attempts this meaningful structural move at this time.”
// Lima evaluates this without mutating anything.
//

#[derive(Debug, Clone)]
pub struct Transition {
    pub domain: DomainId,

    pub actor: ActorId,
    pub seat: SeatId,

    /// Structural move in Terra space.
    pub from: TerraNodeId,
    pub to: TerraNodeId,

    /// Semantic intent (Numen meaning applied to this transition).
    pub meaning: MeaningId,

    /// Time ribbon sample (used for turn-taking, TTLs, freeze windows, etc.)
    pub tick: TickId,
}

impl Transition {
    pub fn new(
        domain: DomainId,
        actor: ActorId,
        seat: SeatId,
        from: TerraNodeId,
        to: TerraNodeId,
        meaning: MeaningId,
        tick: TickId,
    ) -> Self {
        Self { domain, actor, seat, from, to, meaning, tick }
    }
}
