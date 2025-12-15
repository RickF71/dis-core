// src/spine/node.rs

use crate::spine::payload::PayloadRef;
use rand::RngCore;


/// Strong ID type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct NodeId(pub [u8; 32]);

impl NodeId {
    pub fn generate() -> Self {
        // If you already depend on rand, use it. Otherwise swap to something else.
        let mut bytes = [0u8; 32];
        rand::thread_rng().fill_bytes(&mut bytes);
        Self(bytes)
    }
}

#[derive(Debug, Clone)]
pub struct Node {
    pub id: NodeId,
    pub domain: NodeId,
    pub payload: PayloadRef,
}
