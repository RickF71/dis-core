use std::sync::atomic::{AtomicBool, Ordering};
use std::collections::HashMap;
use crate::domain::seat::CorporealSeatProjection;
use crate::domain::domain_id::DomainId;
use crate::spine::clock::DisTick;
use uuid::Uuid;

/// A Seat is persistent and owned.
/// It may be undefined (no actor present), but never empty.
#[derive(Debug)]
pub struct Seat {
    pub id: SeatId,

    /// Root secret (stub for now; opaque bytes)
    secret: SeatSecret,

    /// Whether an actor is currently defining the seat
    actor_present: AtomicBool,

    /// Seat-owned persistent storage
    pub storage: SeatStorage,

    /// Projections owned by this seat
    pub projections: HashMap<DomainId, CorporealSeatProjection>,
}


#[derive(Debug, Clone, Copy)]
pub struct SeatId(pub Uuid);

#[derive(Debug)]
struct SeatSecret([u8; 32]);

impl Seat {
    pub fn new() -> Self {
        Self {
            id: SeatId(Uuid::new_v4()),
            secret: SeatSecret(rand::random()),
            actor_present: AtomicBool::new(false),
            storage: SeatStorage::new(),
            projections: HashMap::new(),
        }
    }
    pub fn upsert_corporeal_projection(
        &mut self,
        domain_id: DomainId,
        dis_tick: DisTick,
    ) {
        let projection = CorporealSeatProjection {
            domain_id,
            owner_seat: self.id,
            seat_domain_tick: 0,
            last_active_dis_tick: dis_tick,
            permissions: crate::domain::seat::DomainPermissions { can_act: true },
        };
        self.projections.insert(domain_id, projection);
    }


    /// Actor arrives: seat becomes defined
    pub fn enter(&self) {
        self.actor_present.store(true, Ordering::SeqCst);
    }

    /// Actor leaves: seat becomes undefined (but persists)
    pub fn leave(&self) {
        self.actor_present.store(false, Ordering::SeqCst);
    }

    pub fn is_defined(&self) -> bool {
        self.actor_present.load(Ordering::SeqCst)
    }
}

#[derive(Debug)]
pub struct SeatStorage {
    pub artifacts: HashMap<ArtifactId, Artifact>,
}

impl SeatStorage {
    pub fn new() -> Self {
        Self {
            artifacts: HashMap::new(),
        }
    }
}

#[derive(Debug, Clone)]
pub struct Artifact {
    pub id: ArtifactId,
    pub created_at: DisTick,
    pub content: String,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct ArtifactId(pub Uuid);
