// src/spine/projection/triad.rs
//
// Canonical Structural Triad
//
// Invariants:
// - Actor is established in Nullus (identity)
// - Seat is established in Aether (continuity / agency channel)
// - Domain is established in Terra (situated structure)
//
// Triad encodes structural placement only.
// It grants no authority and encodes no capability.
// Upper-spine capabilities (Numen / Lima / Corporeal)
// project THROUGH triads elsewhere.

use crate::spine::SpineLayer;

// --- Core identifiers (opaque by design) ---

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct ActorId(pub u128);

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct DomainId(pub u128);

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct SeatId(pub u128);

// --- Structural entities (fact-only) ---

#[derive(Debug, Clone)]
pub struct Actor {
    pub id: ActorId,
}

#[derive(Debug, Clone)]
pub struct Domain {
    pub id: DomainId,
    // Domain == Terra (structure only)
}

#[derive(Debug, Clone)]
pub struct Seat {
    pub id: SeatId,
    pub actor: ActorId,
    pub domain: DomainId,
    // No authority, role, or privilege implied here
}

// --- The Triad ---
/// A Triad represents structural placement only.
/// It is a coincidence of independently established facts.
/// It does not grant authority or permission.
#[derive(Debug, Clone)]
pub struct Triad {
    pub actor: ActorId,
    pub seat: SeatId,
    pub domain: DomainId,
}

impl Triad {
    // These are facts, not configuration.
    pub const ACTOR_LAYER: SpineLayer = SpineLayer::Nullus;
    pub const SEAT_LAYER: SpineLayer = SpineLayer::Aether;
    pub const DOMAIN_LAYER: SpineLayer = SpineLayer::Terra;

    pub fn structural_layers() -> (SpineLayer, SpineLayer, SpineLayer) {
        (Self::ACTOR_LAYER, Self::SEAT_LAYER, Self::DOMAIN_LAYER)
    }
}
