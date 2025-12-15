// src/spine/storage_key.rs
use crate::spine::Layer6;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct StorageKey {
    pub layer: Layer6,
    pub material: [u8; 32],
}
