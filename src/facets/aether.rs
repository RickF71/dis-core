use std::collections::{HashMap, HashSet};
use std::sync::atomic::{AtomicBool, Ordering};

use crate::id::{DomainId, SeatId};
use super::{FacetMeta, FacetKind, FacetId};
use super::interfaces::*;
use super::nullus::NullusFacet;

#[derive(Debug)]
pub struct AetherFacet {
    pub meta: FacetMeta,

    /// Seat registry (persistent, domain-owned)
    seats: HashSet<SeatId>,

    /// Runtime presence flags (seat-definedness)
    present: HashMap<SeatId, AtomicBool>,
}

impl AetherFacet {
    pub fn mint_from_parent(parent: &NullusFacet) -> Self {
        Self {
            meta: FacetMeta::new(parent.meta.domain_id, FacetKind::Aether, Some(parent.meta.id)),
            seats: HashSet::new(),
            present: HashMap::new(),
        }
    }

    pub fn mint_terra(&self) -> super::terra::TerraFacet {
        super::terra::TerraFacet::mint_from_parent(self)
    }
}

impl SeatIfc for AetherFacet {
    fn facet_id(&self) -> FacetId { self.meta.id }
    fn domain_id(&self) -> DomainId { self.meta.domain_id }

    fn mint_seat(&mut self) -> SeatId {
        let id = SeatId(uuid::Uuid::new_v4());
        self.seats.insert(id);
        self.present.insert(id, AtomicBool::new(false));
        id
    }

    fn seat_present(&self, seat: SeatId) -> bool {
        self.present
            .get(&seat)
            .map(|b| b.load(Ordering::SeqCst))
            .unwrap_or(false)
    }

    fn enter_seat(&mut self, seat: SeatId) {
        if let Some(b) = self.present.get(&seat) {
            b.store(true, Ordering::SeqCst);
        }
    }

    fn leave_seat(&mut self, seat: SeatId) {
        if let Some(b) = self.present.get(&seat) {
            b.store(false, Ordering::SeqCst);
        }
    }
}
