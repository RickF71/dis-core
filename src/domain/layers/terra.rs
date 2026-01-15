// src/spine/terra.rs

use crate::spine::Layer6;
use crate::spine::payload::{DomainShape, PayloadShapeSchema, ViewContract};

pub struct TerraShape;

impl DomainShape for TerraShape {
    const DOMAIN: Layer6 = Layer6::Terra;
}

pub struct TerraNullusView;
pub struct TerraAetherView;
pub struct TerraTerraView;
pub struct TerraNumenView;
pub struct TerraLimaView;
pub struct TerraCorporealView;

impl ViewContract for TerraNullusView {
    const LAYER: Layer6 = Layer6::Nullus;
}
impl ViewContract for TerraAetherView {
    const LAYER: Layer6 = Layer6::Aether;
}
impl ViewContract for TerraTerraView {
    const LAYER: Layer6 = Layer6::Terra;
}
impl ViewContract for TerraNumenView {
    const LAYER: Layer6 = Layer6::Numen;
}
impl ViewContract for TerraLimaView {
    const LAYER: Layer6 = Layer6::Lima;
}
impl ViewContract for TerraCorporealView {
    const LAYER: Layer6 = Layer6::Corporeal;
}

impl PayloadShapeSchema for TerraShape {
    type NullusView = TerraNullusView;
    type AetherView = TerraAetherView;
    type TerraView = TerraTerraView;
    type NumenView = TerraNumenView;
    type LimaView = TerraLimaView;
    type CorporealView = TerraCorporealView;
}
