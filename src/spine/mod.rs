// src/spine/mod.rs
//
// Foundational ontology for DIS.
// Defines the Base-6 spine layers, their intrinsic geometry,
// and cross-layer structural projections.
//
// Contains no runtime behavior.

pub mod layer;
pub mod lattice;
pub mod cube;

// Cross-layer projections (fact alignment, no authority)
pub mod projection;

// Layer-local definitions (dimension-specific meaning)
#[path = "01_nullus/mod.rs"]
pub mod nullus;

#[path = "02_aether/mod.rs"]
pub mod aether;

#[path = "03_terra/mod.rs"]
pub mod terra;

#[path = "04_numen/mod.rs"]
pub mod numen;

#[path = "05_lima/mod.rs"]
pub mod lima;

#[path = "06_corporeal/mod.rs"]
pub mod corporeal;

// --- Re-exports: ontology ---
pub use layer::SpineLayer;
pub use lattice::{
    SpineLattice,
    SpineError,
    validate_lattice_stack,
    validate_full_spine,
};
pub use cube::SpineCube;

// --- Re-exports: projections ---
pub use projection::triad::{
    ActorId,
    DomainId,
    SeatId,
    Actor,
    Domain,
    Seat,
    Triad,
};
