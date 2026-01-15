use std::collections::{HashMap, HashSet};

use crate::spine::terra::{TerraNodeId, TerraEdgeId};

#[derive(Debug, Default)]
pub struct TerraGraph {
    pub nodes: HashMap<TerraNodeId, ()>,
    pub edges: HashMap<TerraNodeId, HashSet<TerraEdgeId>>,
}
