use std::collections::HashMap;
use std::sync::atomic::{AtomicBool, Ordering};

use uuid::Uuid;

use crate::domain::domain_id::DomainId;
use crate::domain::seat::CorporealSeatProjection;
use crate::spine::clock::DisTick;
use crate::domain::parent_link::ParentLink;
use crate::domain::reflective_surface::ReflectiveSurface;


/// A Seat is persistent and owned.
/// It may be undefined (no actor present), but it is never empty.
///
/// A Seat does NOT contain intrinsic secrets.
/// All authorization material is stored in seat-owned storage,
/// scoped to external (target) domains.
#[derive(Debug)]
pub struct Seat {
    /// Stable seat identifier
    pub id: SeatId,

    /// Whether an actor is currently defining the seat
    actor_present: AtomicBool,

    /// Seat-owned persistent storage (totem DB)
    pub storage: SeatStorage,

    /// Live projections owned by this seat (operational presence)
    pub projections: HashMap<DomainId, CorporealSeatProjection>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct SeatId(pub Uuid);

impl Seat {
    pub fn new() -> Self {
        Self {
            id: SeatId(Uuid::new_v4()),
            actor_present: AtomicBool::new(false),
            storage: SeatStorage::new(),
            projections: HashMap::new(),
        }
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

    /// Create or update a corporeal projection for a target domain
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
            permissions: crate::domain::seat::DomainPermissions {
                can_act: true,
            },
        };

        self.projections.insert(domain_id, projection);
    }
}

/// Seat-owned persistent storage.
///
/// This is the user's private organizational and authorization space.
/// Nothing here is authoritative outside the seat's domain.
#[derive(Debug)]
pub struct SeatStorage {
    /// User artifacts (notes, receipts, cached objects, etc.)
    pub artifacts: HashMap<ArtifactId, Artifact>,

    /// Virtual domains representing "who I am where"
    pub virtual_domains: HashMap<DomainId, VirtualDomain>,
}

impl SeatStorage {
    pub fn new() -> Self {
        Self {
            artifacts: HashMap::new(),
            virtual_domains: HashMap::new(),
        }
    }
}

/// A virtual domain is a private, user-maintained projection
/// of an external domain.
///
/// It stores the minimum authorization material (a secret)
/// plus any personal context needed to make repeated interaction sane.
#[derive(Debug)]
pub struct VirtualDomain {
    /// Target (external) domain
    pub target_domain: DomainId,

    /// How I speak to this domain (entry capability)
    pub entry_secret: EntrySecret,

    /// How this domain is anchored upward (lineage)
    pub parent_link: ParentLink,

    /// How I believe this domain sees/constrains me
    pub reflective_surface: ReflectiveSurface,
}


/// Domain-scoped authorization secret.
/// This is NOT an identity secret.
#[derive(Debug)]
pub struct EntrySecret(pub [u8; 32]);

#[derive(Debug, Clone)]
pub struct Artifact {
    pub id: ArtifactId,
    pub created_at: DisTick,
    pub content: String,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct ArtifactId(pub Uuid);
