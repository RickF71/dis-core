// src/spine/05_lima/reasons.rs
use crate::spine::lima::types::*;

#[derive(Debug, Clone)]
pub enum DenyReason {
    // --------------------------------
    // Aether / presence
    // --------------------------------
    SeatNotBound { seat: SeatId },
    ActorNotOccupant { seat: SeatId, actor: ActorId },

    // --------------------------------
    // Terra / structure
    // --------------------------------
    NoStructuralPath { from: TerraNodeId, to: TerraNodeId },

    // --------------------------------
    // Numen / meaning
    // --------------------------------
    MeaningNotApplicable { meaning: MeaningId },

    // Optional: contracts not satisfied (Numen describes them; Lima checks preconditions)
    ContractNotSatisfied { contract: ContractId },

    // --------------------------------
    // Time / turn / sequencing
    // --------------------------------
    NotYourTurn { seat: SeatId },

    // --------------------------------
    // Freeze ribbon
    // --------------------------------
    Frozen { scope: FreezeScopeId },

    // --------------------------------
    // Catch-all
    // --------------------------------
    Other { code: String, detail: String },
}
