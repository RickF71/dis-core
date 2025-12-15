pub mod clock;
pub mod capsule;
pub mod node;
pub mod payload;
pub mod echo_store;
pub mod spine;
pub mod phases;



use serde::Serialize;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize)]
pub enum Layer6 {
    Nullus,
    Aether,
    Terra,
    Numen,
    Lima,
    Corporeal,
}

