use crate::domain::lattice::DomainLattice;
use uuid::Uuid;

#[derive(Debug, Clone)]
pub struct Artifact {
    pub id: ArtifactId,
    pub owner_seat: SeatId,
    pub created_at: DomainLattice,
    pub content: String,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct ArtifactId(pub Uuid);

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct SeatId(pub Uuid);
