// src/runtime/domain_runtime.rs
//! DomainRuntime
//!
//! A DomainRuntime represents a *running domain*.
//! It contains the six aspects as live components,
//! but does not itself define policy, storage, or simulation.
//!
//! This is a coordination shell, not an authority.
//! //! IMPORTANT:
//! DomainRuntime is not authoritative.
//! All authority, meaning, and validity are defined
//! outside the runtime and projected through it.


use crate::domain::domain::Domain;
use crate::id::DomainId;

pub struct DomainRuntime {
    /// Canonical domain ontology (includes identity)
    pub domain: Domain,

    // --- Running aspects (placeholders for now) ---

    /// Identity & inscription (existence anchoring)
    pub nullus: NullusRuntime,

    /// Memory & continuity
    pub aether: AetherRuntime,

    /// Structure & topology
    pub terra: TerraRuntime,

    /// Meaning traversal (runtime only)
    pub numen: NumenRuntime,

    /// Understanding / presentation of memory
    pub lima: LimaRuntime,

    /// External interface / projection
    pub corporeal: CorporealRuntime,
}

// --- Empty runtime aspect shells ---
// These intentionally do nothing yet.

pub struct NullusRuntime;
pub struct AetherRuntime;
pub struct TerraRuntime;
pub struct NumenRuntime;
pub struct LimaRuntime;
pub struct CorporealRuntime;

impl DomainRuntime {
    /// Create a new running domain shell.
    ///
    /// This does NOT:
    /// - allocate storage
    /// - start threads
    /// - open sockets
    /// - perform validation
    ///
    /// It only establishes that the domain *exists as a running structure*.
    pub fn new(domain: Domain) -> Self {
        Self {
            domain,
            nullus: NullusRuntime,
            aether: AetherRuntime,
            terra: TerraRuntime,
            numen: NumenRuntime,
            lima: LimaRuntime,
            corporeal: CorporealRuntime,
        }
    }

    /// Access the domain's intrinsic identity.
    pub fn id(&self) -> &DomainId {
        &self.domain.id
    }
}
