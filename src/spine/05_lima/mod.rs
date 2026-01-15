// src/spine/05_lima/mod.rs
//
// Lima — Constraint / Commitment Layer (D5)
//
// Evaluates proposed transitions (hypotheses) against coherence constraints.
// NO state mutation.
// NO rendering.
// NO receipt writing.
//
// Lima consumes:
// - Aether (presence / bindings)
// - Terra (structure / adjacency)
// - Numen (meaning / contracts)
// - Time ribbon (ticks)
// - Freeze ribbon (deny scopes)
//
// Lima produces:
// - Decision (Allow / Deny + reason)
//
// NOTE: This module is intentionally domain-agnostic.
// Domains plug in via the TerraView / NumenView traits and optional DomainRules.

pub mod types;
pub mod transition;
pub mod decision;
pub mod reasons;
pub mod context;
pub mod evaluator;
pub mod receipt;


pub use types::*;
pub use transition::*;
pub use decision::*;
pub use reasons::*;
pub use context::*;
pub use evaluator::*;
pub use receipt::TransitionReceipt;


