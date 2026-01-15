use std::collections::HashMap;
use std::sync::atomic::{AtomicBool, Ordering};

use crate::domain::parent_link::ParentLink;
use crate::domain::lattice::DomainLattice;
use crate::domain::lattice_axis::LatticeAxis;
use crate::domain::artifact::Artifact;
use crate::id::{ArtifactId, DomainId, SeatId};

/// ----------------------------
/// Prime Seat (authority anchor)
/// ----------------------------
///
/// A PrimeSeat is NEVER self-created.
/// It is minted by a parent domain's PrimeSeat.
/// It anchors authority and authorizes lattice actions.
#[derive(Debug)]
pub struct PrimeSeat {
    pub id: SeatId,

    /// Lineage: who minted this prime seat
    pub parent: ParentLink,

    /// Challenge secret used to derive child authority
    pub challenge_secret: ChallengeSecret,
}

impl PrimeSeat {
    /// Constructor is crate-private to prevent self-minting
    pub(crate) fn mint_from_parent(parent: ParentLink) -> Self {
        Self {
            id: SeatId(uuid::Uuid::new_v4()),
            parent,
            challenge_secret: ChallengeSecret::new(),
        }
    }

    /// Authorize an action along a lattice axis.
    /// Policy enforcement will live above this.
    pub fn authorize_axis(&self, _axis: LatticeAxis) -> bool {
        true
    }
}

/// ----------------------------
/// Generic Seat (user presence)
/// ----------------------------
///
/// A Seat is persistent and owned.
/// It may be undefined (no actor present),
/// but it is never empty or authority-bearing.
#[derive(Debug)]
pub struct Seat {
    pub id: SeatId,

    /// Runtime presence (actor may come and go)
    actor_present: AtomicBool,

    /// Seat-owned private storage (non-authoritative)
    pub storage: SeatStorage,

    /// Corporeal projections into external domains
    pub projections: HashMap<DomainId, CorporealSeatProjection>,
}

impl Seat {
    /// Seats are created only by domain logic.
    /// This constructor is crate-private.
    pub(crate) fn new() -> Self {
        Self {
            id: SeatId(uuid::Uuid::new_v4()),
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
}

/// ----------------------------
/// Corporeal Seat Projection
/// ----------------------------
///
/// A projection is NOT authority.
/// It is permissioned presence inside another domain.
#[derive(Debug, Clone)]
pub struct CorporealSeatProjection {
    /// Domain this projection exists in
    pub domain_id: DomainId,

    /// Owning seat (never leaves home domain)
    pub owner_seat: SeatId,

    /// Highest lattice coordinate observed in that domain
    pub observed_lattice: DomainLattice,

    /// Domain-granted permissions (opaque)
    pub permissions: DomainPermissions,
}

#[derive(Debug, Clone)]
pub struct DomainPermissions {
    pub can_act: bool,
}

/// ----------------------------
/// Seat-Owned Private Storage
/// ----------------------------
///
/// This storage is personal, local, and non-authoritative.
/// Nothing here conveys power outside the seat's home domain.
#[derive(Debug)]
pub struct SeatStorage {
    pub artifacts: HashMap<ArtifactId, Artifact>,
}

impl SeatStorage {
    fn new() -> Self {
        Self {
            artifacts: HashMap::new(),
        }
    }
}

/// ----------------------------
/// Secrets
/// ----------------------------

#[derive(Debug)]
pub struct ChallengeSecret([u8; 32]);

impl ChallengeSecret {
    fn new() -> Self {
        use rand::RngCore;
        let mut bytes = [0u8; 32];
        rand::thread_rng().fill_bytes(&mut bytes);
        Self(bytes)
    }
}
