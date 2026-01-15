// src/taiji/snapshot.rs

use serde::Serialize;
use crate::spine::SpineCube;

// ----------------------------
// Observer-facing views
// ----------------------------

#[derive(Debug, Clone, Serialize)]
pub struct NullusView {
    pub actor_present: bool,
}

#[derive(Debug, Clone, Serialize)]
pub struct AetherView {
    pub seat_present: bool,
}

#[derive(Debug, Clone, Serialize)]
pub struct TerraView {
    pub domain_id: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct NumenView {
    pub authority_visible: bool,
}

#[derive(Debug, Clone, Serialize)]
pub struct LimaView {
    pub intent_open: bool,
}

#[derive(Debug, Clone, Serialize)]
pub struct CorporealView {
    pub last_transition: Option<String>,
}

// ----------------------------
// Canonical Taiji snapshot
// ----------------------------

#[derive(Debug, Clone, Serialize)]
pub struct Snapshot6D {
    /// Witnessed tick (starts at 1)
    pub sequence: u64,

    /// One 6-face cube of observed reality
    pub cube: SpineCube<
        NullusView,
        AetherView,
        TerraView,
        NumenView,
        LimaView,
        CorporealView,
    >,
}
