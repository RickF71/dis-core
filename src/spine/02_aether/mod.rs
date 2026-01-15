// src/spine/02_aether/mod.rs
//
// Aether — Continuity Layer (D2)

use std::collections::HashMap;
use crate::spine::projection::triad::{ActorId, DomainId, SeatId};

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct SeatKindId(pub String);

#[derive(Debug, Clone)]
pub struct AetherSeat {
    pub id: SeatId,
    pub domain: DomainId,
    pub kind: SeatKindId,
    occupant: Option<ActorId>,
}

impl AetherSeat {
    pub fn new(id: SeatId, domain: DomainId, kind: SeatKindId) -> Self {
        Self {
            id,
            domain,
            kind,
            occupant: None,
        }
    }

    pub fn occupant(&self) -> Option<&ActorId> {
        self.occupant.as_ref()
    }

    // --- RECEIPT APPLICATION ONLY ---
    pub(crate) fn apply_bind(&mut self, actor: ActorId) {
        self.occupant = Some(actor);
    }

    pub(crate) fn apply_unbind(&mut self) {
        self.occupant = None;
    }
}

#[derive(Debug, Default)]
pub struct AetherState {
    seats: HashMap<SeatId, AetherSeat>,
}

impl AetherState {
    pub fn insert_seat(&mut self, seat: AetherSeat) {
        self.seats.insert(seat.id.clone(), seat);
    }

    // --- RECEIPT APPLICATION ONLY ---
    pub(crate) fn apply_bind(&mut self, seat: &SeatId, actor: ActorId) {
        if let Some(s) = self.seats.get_mut(seat) {
            s.apply_bind(actor);
        }
    }

    pub(crate) fn apply_unbind(&mut self, seat: &SeatId) {
        if let Some(s) = self.seats.get_mut(seat) {
            s.apply_unbind();
        }
    }
}

// ------------------------------------------------------------
// AetherView — Read-only continuity interface
// ------------------------------------------------------------

pub trait AetherView {
    fn occupant(&self, seat: &SeatId) -> Option<&ActorId>;
}

impl AetherView for AetherState {
    fn occupant(&self, seat: &SeatId) -> Option<&ActorId> {
        self.seats.get(seat)?.occupant()
    }
}
