// src/spine/01_nullus/mod.rs
//
// Nullus — Identity Layer (D1)
//
// Responsibility:
// - Establishes the existence of an Actor
// - Guarantees identity continuity
//
// Invariants:
// - Exactly one Nullus tick is consumed per establishment
// - No domains, seats, roles, authority, or meaning exist here
// - Nullus knows NOTHING about higher layers

use crate::spine::projection::triad::ActorId;

/// A Nullus-level actor.
///
/// This represents *identity existence only*.
/// It implies no agency, authority, role, or placement.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct NullusActor {
    pub id: ActorId,
}

impl NullusActor {
    /// Establish a new actor identity (genesis).
    ///
    /// Consumes exactly one Nullus tick.
    /// Elevation beyond Nullus is not implied.
    pub fn create(id: ActorId) -> Self {
        Self { id }
    }

    /// Consume / recognize an existing actor identity.
    ///
    /// Used for re-entry, restoration, or continuity.
    /// Also consumes exactly one Nullus tick.
    pub fn consume(id: ActorId) -> Self {
        Self { id }
    }
}
