// src/spine/lima.rs

use crate::spine::Layer6;
use crate::spine::payload::{DomainShape, PayloadShapeSchema, ViewContract};

pub struct LimaShape;

impl DomainShape for LimaShape {
    const DOMAIN: Layer6 = Layer6::Lima;
}

pub struct LimaNullusView;
pub struct LimaAetherView;
pub struct LimaTerraView;
pub struct LimaNumenView;
pub struct LimaLimaView;
pub struct LimaCorporealView;

impl ViewContract for LimaNullusView {
    const LAYER: Layer6 = Layer6::Nullus;
}
impl ViewContract for LimaAetherView {
    const LAYER: Layer6 = Layer6::Aether;
}
impl ViewContract for LimaTerraView {
    const LAYER: Layer6 = Layer6::Terra;
}
impl ViewContract for LimaNumenView {
    const LAYER: Layer6 = Layer6::Numen;
}
impl ViewContract for LimaLimaView {
    const LAYER: Layer6 = Layer6::Lima;
}
impl ViewContract for LimaCorporealView {
    const LAYER: Layer6 = Layer6::Corporeal;
}

impl PayloadShapeSchema for LimaShape {
    type NullusView = LimaNullusView;
    type AetherView = LimaAetherView;
    type TerraView = LimaTerraView;
    type NumenView = LimaNumenView;
    type LimaView = LimaLimaView;
    type CorporealView = LimaCorporealView;
}
