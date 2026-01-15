// src/spine/03_terra/node.rs
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct TerraNodeId(pub String);

#[derive(Debug, Clone)]
pub struct TerraNode {
    pub id: TerraNodeId,
}
