// src/lib.rs (or wherever that failing test is)
pub mod domain;
pub mod spine;

use crate::spine::Layer6;
use crate::spine::payload::{PayloadRef, PayloadId};
use crate::spine::node::{Node, NodeId};

#[test]
fn builds_node() {
    let id = NodeId::generate();

    let node = Node {
        id,
        domain: id,              // pick your intended default; this is “self-domain”
        payload: PayloadRef {
            domain: Layer6::Nullus,
            id: PayloadId([0u8; 32]),
        }
    };
    // assert something if you want
    let _ = node;
}


