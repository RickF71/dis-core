// src/spine/03_terra/mod.rs
//
// Spine layer: terra
// Dimension: 03
//
// This module defines layer-local concepts ONLY.
// No elevation, no ticks, no authority leakage.
// Runtime behavior lives elsewhere.

pub mod node;
pub mod edge;
pub mod graph;

pub use node::TerraNodeId;
pub use edge::TerraEdgeId;


use crate::spine::projection::triad::{DomainId, SeatId};



#[derive(Debug, Clone)]
pub struct TerraNode {
    pub id: TerraNodeId,
    pub domain: DomainId,

    /// Optional seat anchored at this node
    pub seat: Option<SeatId>,
}

#[derive(Debug, Clone)]
pub struct TerraEdge {
    pub id: TerraEdgeId,
    pub from: TerraNodeId,
    pub to: TerraNodeId,

    /// Domain-defined structural relation
    pub kind: String,
}

use std::collections::{HashMap, HashSet};

#[derive(Debug, Default)]
pub struct TerraGraph {
    pub nodes: HashMap<TerraNodeId, TerraNode>,
    pub edges: HashMap<TerraNodeId, HashSet<TerraEdge>>,
}
