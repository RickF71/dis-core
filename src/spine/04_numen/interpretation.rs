use std::collections::HashSet;

use crate::spine::terra::node::TerraNodeId;
use crate::spine::terra::edge::TerraEdgeId;

use super::meaning::MeaningId;
use super::contract::ContractId;

#[derive(Debug, Clone)]
pub struct Interpretation {
    pub target_meaning: MeaningId,

    /// Structural elements this meaning applies to
    pub nodes: HashSet<TerraNodeId>,
    pub edges: HashSet<TerraEdgeId>,

    /// Semantic contracts associated with this meaning
    pub contracts: HashSet<ContractId>,
}
