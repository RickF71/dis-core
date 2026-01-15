// src/spine/05_lima/types.rs
//
// Centralizes type aliases to keep Lima stable.
// Adjust the module paths here if your Terra/Numen modules use different names/paths.

/// --- Projection IDs ---
pub type ActorId = crate::spine::projection::triad::ActorId;
pub type DomainId = crate::spine::projection::triad::DomainId;
pub type SeatId  = crate::spine::projection::triad::SeatId;

/// --- Terra IDs (adjust these paths to match your Terra module) ---
/// e.g. crate::spine::03_terra::node::TerraNodeId
pub type TerraNodeId = crate::spine::terra::node::TerraNodeId;
pub type TerraEdgeId = crate::spine::terra::edge::TerraEdgeId;

/// --- Numen IDs (adjust these paths to match your Numen module) ---
pub type MeaningId   = crate::spine::numen::meaning::MeaningId;
pub type ContractId  = crate::spine::numen::contract::ContractId;

/// --- Lima-local IDs ---
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct TickId(pub u64);

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct FreezeScopeId(pub String);
