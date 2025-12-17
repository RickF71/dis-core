// src/lib.rs

pub mod domain;
pub mod spine;

#[cfg(test)]
mod tests {
    //use super::*; // brings `domain` and `spine` into scope

    use crate::spine::Layer6;
    use crate::spine::payload::{PayloadRef, PayloadId};
    use crate::spine::node::{Node, NodeId};

    #[test]
    fn builds_node() {
        let id = NodeId::generate();

        let node = Node {
            id,
            domain: id, // self-domain
            payload: PayloadRef {
                domain: Layer6::Nullus,
                id: PayloadId([0u8; 32]),
            },
        };

        let _ = node;
    }
}
