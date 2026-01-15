use crate::spine::projection::triad::{ActorId, DomainId, SeatId};
use crate::spine::terra::TerraNodeId;
use crate::spine::numen::MeaningId;

#[derive(Debug, Clone)]
pub struct CorporealIntent {
    pub domain: DomainId,
    pub actor: ActorId,
    pub seat: SeatId,

    pub from: TerraNodeId,
    pub to: TerraNodeId,

    pub meaning: MeaningId,

    /// Optional raw input (mouse, touch, key, voice)
    pub source: Option<String>,
}
