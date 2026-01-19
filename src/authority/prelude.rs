// ============================================================
// FILE: src/authority/prelude.rs
// ============================================================

// Intentionally tiny prelude for modules that must touch authority.
// Do NOT re-export internal modules.
pub use super::types::*;
pub use super::errors::{AuthorityError, DenyReason};
pub use super::gate::{AuthorityKernel, AuthorityKernelConfig};

