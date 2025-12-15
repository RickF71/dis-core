// src/spine/corporeal.rs

use crate::domain::seat::Seat;
use crate::domain::domain_id::DomainId;
use crate::spine::clock::DisTick;
use crate::spine::Layer6;
use crate::spine::payload::{DomainShape, PayloadShapeSchema, ViewContract};

pub struct CorporealShape;

impl DomainShape for CorporealShape {
    const DOMAIN: Layer6 = Layer6::Corporeal;
}

pub struct CorporealNullusView;
pub struct CorporealAetherView;
pub struct CorporealTerraView;
pub struct CorporealNumenView;
pub struct CorporealLimaView;
pub struct CorporealCorporealView;

impl ViewContract for CorporealNullusView {
    const LAYER: Layer6 = Layer6::Nullus;
}
impl ViewContract for CorporealAetherView {
    const LAYER: Layer6 = Layer6::Aether;
}
impl ViewContract for CorporealTerraView {
    const LAYER: Layer6 = Layer6::Terra;
}
impl ViewContract for CorporealNumenView {
    const LAYER: Layer6 = Layer6::Numen;
}
impl ViewContract for CorporealLimaView {
    const LAYER: Layer6 = Layer6::Lima;
}
impl ViewContract for CorporealCorporealView {
    const LAYER: Layer6 = Layer6::Corporeal;
}

impl PayloadShapeSchema for CorporealShape {
    type NullusView = CorporealNullusView;
    type AetherView = CorporealAetherView;
    type TerraView = CorporealTerraView;
    type NumenView = CorporealNumenView;
    type LimaView = CorporealLimaView;
    type CorporealView = CorporealCorporealView;
}

// Example execution point for Corporeal phase:
fn execute_corporeal_phase(seat: &mut Seat, dis_tick: DisTick) {
    seat.upsert_corporeal_projection(
        DomainId::self_domain(),
        dis_tick,
    );
}
