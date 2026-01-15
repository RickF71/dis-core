// src/spine/elevation.rs
//
// Canonical Elevation Contract
//
// Elevation transforms a TotPoint (D1) into a TotSphere (D6)
// by passing deterministically through all six spine layers.
//
// Invariants:
// - Exactly one tick is consumed per layer
// - Exactly one ribbon is added per layer
// - Each layer adds exactly one degree of freedom
// - ActorId continuity is preserved across all layers
//
// This file defines STRUCTURE ONLY.
// No authority, no policy, no mutation logic.

use crate::spine::SpineLayer;
use crate::spine::projection::triad::ActorId;

// ------------------------------------------------------------
// Core Tick + Ribbon
// ------------------------------------------------------------

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Tick(pub u64);

#[derive(Debug, Clone)]
pub struct Ribbon {
    pub layer: SpineLayer,
    pub actor: ActorId,

    // Capability: this layer may spawn children at its level.
    // This is NOT authority, only structural permission.
    pub can_spawn: bool,
}

// ------------------------------------------------------------
// TotPoint (D1)
// ------------------------------------------------------------

#[derive(Debug, Clone)]
pub struct TotPoint {
    pub actor: ActorId,
    pub tick_nullus: Tick,
    pub ribbons: Vec<Ribbon>,
}

impl TotPoint {
    pub fn new(actor: ActorId) -> Self {
        Self {
            actor,
            tick_nullus: Tick(1), // Nullus consumes tick 0 → 1
            ribbons: vec![Ribbon {
                layer: SpineLayer::Nullus,
                actor,
                can_spawn: true,
            }],
        }
    }
}

// ------------------------------------------------------------
// TotSphere (D6)
// ------------------------------------------------------------

#[derive(Debug, Clone)]
pub struct TotSphere {
    pub actor: ActorId,

    // Tick vector: always length 6
    pub ticks: [Tick; 6],

    // Ribbons: exactly one per layer, ordered
    pub ribbons: Vec<Ribbon>,
}

// ------------------------------------------------------------
// Layer-local elevation outputs (opaque by design)
// ------------------------------------------------------------

pub struct NullusOut;
pub struct AetherOut;
pub struct TerraOut;
pub struct NumenOut;
pub struct LimaOut;
pub struct CorporealOut;

// ------------------------------------------------------------
// Elevation Traits (one per layer)
// ------------------------------------------------------------

pub trait ElevateNullus {
    fn elevate_from_nullus(point: TotPoint) -> (AetherOut, Tick, Ribbon);
}

pub trait ElevateAether {
    fn elevate_from_aether(
        actor: ActorId,
        prev: AetherOut,
    ) -> (TerraOut, Tick, Ribbon);
}

pub trait ElevateTerra {
    fn elevate_from_terra(
        actor: ActorId,
        prev: TerraOut,
    ) -> (NumenOut, Tick, Ribbon);
}

pub trait ElevateNumen {
    fn elevate_from_numen(
        actor: ActorId,
        prev: NumenOut,
    ) -> (LimaOut, Tick, Ribbon);
}

pub trait ElevateLima {
    fn elevate_from_lima(
        actor: ActorId,
        prev: LimaOut,
    ) -> (CorporealOut, Tick, Ribbon);
}

// ------------------------------------------------------------
// Canonical Full Elevation Composition
// ------------------------------------------------------------

pub trait ElevateAll:
    ElevateNullus
    + ElevateAether
    + ElevateTerra
    + ElevateNumen
    + ElevateLima
{
    fn elevate_all(point: TotPoint) -> TotSphere {
        let actor = point.actor;

        // Nullus → Aether
        let (aether, t2, r2) =
            <Self as ElevateNullus>::elevate_from_nullus(point);

        // Aether → Terra
        let (terra, t3, r3) =
            <Self as ElevateAether>::elevate_from_aether(actor, aether);

        // Terra → Numen
        let (numen, t4, r4) =
            <Self as ElevateTerra>::elevate_from_terra(actor, terra);

        // Numen → Lima
        let (lima, t5, r5) =
            <Self as ElevateNumen>::elevate_from_numen(actor, numen);

        // Lima → Corporeal
        let (_corp, t6, r6) =
            <Self as ElevateLima>::elevate_from_lima(actor, lima);

        TotSphere {
            actor,
            ticks: [
                Tick(1), // Nullus always ends at 1 on initialization
                t2,
                t3,
                t4,
                t5,
                t6,
            ],
            ribbons: vec![
                Ribbon {
                    layer: SpineLayer::Nullus,
                    actor,
                    can_spawn: true,
                },
                r2,
                r3,
                r4,
                r5,
                r6,
            ],
        }
    }
}
