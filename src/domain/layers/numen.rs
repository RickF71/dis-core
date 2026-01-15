// src/spine/numen.rs

use crate::spine::Layer6;
use crate::spine::payload::{DomainShape, PayloadShapeSchema, ViewContract};

pub struct NumenShape;

impl DomainShape for NumenShape {
    const DOMAIN: Layer6 = Layer6::Numen;
}

pub struct NumenNullusView;
pub struct NumenAetherView;
pub struct NumenTerraView;
pub struct NumenNumenView;
pub struct NumenLimaView;
pub struct NumenCorporealView;

impl ViewContract for NumenNullusView {
    const LAYER: Layer6 = Layer6::Nullus;
}
impl ViewContract for NumenAetherView {
    const LAYER: Layer6 = Layer6::Aether;
}
impl ViewContract for NumenTerraView {
    const LAYER: Layer6 = Layer6::Terra;
}
impl ViewContract for NumenNumenView {
    const LAYER: Layer6 = Layer6::Numen;
}
impl ViewContract for NumenLimaView {
    const LAYER: Layer6 = Layer6::Lima;
}
impl ViewContract for NumenCorporealView {
    const LAYER: Layer6 = Layer6::Corporeal;
}

impl PayloadShapeSchema for NumenShape {
    type NullusView = NumenNullusView;
    type AetherView = NumenAetherView;
    type TerraView = NumenTerraView;
    type NumenView = NumenNumenView;
    type LimaView = NumenLimaView;
    type CorporealView = NumenCorporealView;
}
