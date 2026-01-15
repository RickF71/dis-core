pub mod kinds;
pub mod ids;
pub mod interfaces;

pub mod nullus;
pub mod aether;
pub mod terra;
pub mod numen;
pub mod lima;
pub mod corporeal;

pub mod domain_stack;

pub use kinds::FacetKind;
pub use ids::{FacetId, FacetMeta};
pub use interfaces::*;
pub use domain_stack::DomainStack;
