// src/domain/artifact/artifact.rs

use crate::domain::lattice::DomainLattice;
use crate::id::{ArtifactId, SeatId};

#[derive(Debug, Clone)]
pub struct Artifact {
    pub id: ArtifactId,
    pub owner_seat: SeatId,
    pub created_at: DomainLattice,
    pub content: String,
}
