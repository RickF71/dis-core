// src/spine/03_terra/edge.rs
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct TerraEdgeId(pub String);

#[derive(Debug, Clone)]
pub struct TerraEdge {
    pub from: super::TerraNodeId,
    pub to: super::TerraNodeId,
}
