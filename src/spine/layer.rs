// src/spine/layer.rs
//
// DIS Spine Layers (Base-6)
// Think: 6 fixed faces of one cube (static ontology).
//
// These are semantic layers, not runtime modules.

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub enum SpineLayer {
    Nullus = 1,
    Aether = 2,
    Terra  = 3,
    Numen  = 4,
    Lima   = 5,
    Corporeal = 6,
}

impl SpineLayer {
    pub const ALL: [SpineLayer; 6] = [
        SpineLayer::Nullus,
        SpineLayer::Aether,
        SpineLayer::Terra,
        SpineLayer::Numen,
        SpineLayer::Lima,
        SpineLayer::Corporeal,
    ];
}
