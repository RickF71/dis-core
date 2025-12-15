use crate::spine::clock::DisTick;
use uuid::Uuid;

#[derive(Debug, Clone)]
pub struct Artifact {
    pub id: ArtifactId,
    pub owner_seat: SeatId,
    pub created_at: DisTick,
    pub content: String,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct ArtifactId(pub Uuid);

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct SeatId(pub Uuid);
