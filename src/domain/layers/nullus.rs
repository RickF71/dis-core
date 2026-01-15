// src/spine/nullus.rs

use crate::spine::Layer6;
use crate::spine::payload::{DomainShape, PayloadShapeSchema, ViewContract};

/// Payload shape for the Nullus domain.
/// Nullus acknowledges existence but does not add meaning.
pub struct NullusShape;

impl DomainShape for NullusShape {
    const DOMAIN: Layer6 = Layer6::Nullus;
}

// ---- View structs ----
// These exist by structural law, even if most are inert.

pub struct NullusNullusView;
pub struct NullusAetherView;
pub struct NullusTerraView;
pub struct NullusNumenView;
pub struct NullusLimaView;
pub struct NullusCorporealView;

// ---- ViewContract implementations ----

impl ViewContract for NullusNullusView {
    const LAYER: Layer6 = Layer6::Nullus;
}

impl ViewContract for NullusAetherView {
    const LAYER: Layer6 = Layer6::Aether;
}

impl ViewContract for NullusTerraView {
    const LAYER: Layer6 = Layer6::Terra;
}

impl ViewContract for NullusNumenView {
    const LAYER: Layer6 = Layer6::Numen;
}

impl ViewContract for NullusLimaView {
    const LAYER: Layer6 = Layer6::Lima;
}

impl ViewContract for NullusCorporealView {
    const LAYER: Layer6 = Layer6::Corporeal;
}

// ---- Payload shape schema ----
// Nullus defines a full projection grid, even if behavior is minimal.

impl PayloadShapeSchema for NullusShape {
    type NullusView = NullusNullusView;
    type AetherView = NullusAetherView;
    type TerraView = NullusTerraView;
    type NumenView = NullusNumenView;
    type LimaView = NullusLimaView;
    type CorporealView = NullusCorporealView;
}
