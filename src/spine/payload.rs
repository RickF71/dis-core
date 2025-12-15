// src/spine/payload.rs

use std::marker::PhantomData;
use crate::spine::Layer6;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct PayloadId(pub [u8; 32]);

#[derive(Debug, Clone, Copy)]
pub struct PayloadRef {
    pub domain: Layer6,
    pub id: PayloadId,
}

// Marker trait — intentionally empty
pub trait DomainShape {
    const DOMAIN: Layer6;
}

// Domain-shaped payload
pub struct Payload<S: PayloadShapeSchema> {
    pub id: PayloadId,
    _shape: PhantomData<S>,
}

pub trait ViewContract {
    const LAYER: Layer6;
}

pub trait PayloadShapeSchema: DomainShape {
    type NullusView: ViewContract;
    type AetherView: ViewContract;
    type TerraView: ViewContract;
    type NumenView: ViewContract;
    type LimaView: ViewContract;
    type CorporealView: ViewContract;
}

