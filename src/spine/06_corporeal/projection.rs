use crate::spine::terra::TerraNodeId;

#[derive(Debug, Clone)]
pub struct CorporealProjection {
    pub visible_nodes: Vec<TerraNodeId>,
    pub highlighted_nodes: Vec<TerraNodeId>,
}
