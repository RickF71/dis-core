// src/spine/aether.rs

use crate::spine::Layer6;
use crate::spine::payload::{DomainShape, PayloadShapeSchema, ViewContract};

pub struct AetherShape;

impl DomainShape for AetherShape {
    const DOMAIN: Layer6 = Layer6::Aether;
}

pub struct AetherNullusView;
pub struct AetherAetherView;
pub struct AetherTerraView;
pub struct AetherNumenView;
pub struct AetherLimaView;
pub struct AetherCorporealView;

impl ViewContract for AetherNullusView {
    const LAYER: Layer6 = Layer6::Nullus;
}
impl ViewContract for AetherAetherView {
    const LAYER: Layer6 = Layer6::Aether;
}
impl ViewContract for AetherTerraView {
    const LAYER: Layer6 = Layer6::Terra;
}
impl ViewContract for AetherNumenView {
    const LAYER: Layer6 = Layer6::Numen;
}
impl ViewContract for AetherLimaView {
    const LAYER: Layer6 = Layer6::Lima;
}
impl ViewContract for AetherCorporealView {
    const LAYER: Layer6 = Layer6::Corporeal;
}

impl PayloadShapeSchema for AetherShape {
    type NullusView = AetherNullusView;
    type AetherView = AetherAetherView;
    type TerraView = AetherTerraView;
    type NumenView = AetherNumenView;
    type LimaView = AetherLimaView;
    type CorporealView = AetherCorporealView;
}
