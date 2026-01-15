// src/runtime/node_runtime.rs
//! NodeRuntime
//!
//! A NodeRuntime is a process-local host for multiple running domains.
//!
//! Responsibilities:
//! - Own the lifetime of DomainRuntime instances
//! - Provide lookup by DomainId
//! - Serve as the attachment point for WS / API / adapters
//!
//! Non-responsibilities:
//! - No authority or policy decisions
//! - No storage, persistence, or simulation
//! - No networking or transport logic
//!
//! NodeRuntime is coordination only.

use std::collections::HashMap;

use crate::domain::Domain;
use crate::id::DomainId;

use super::domain_runtime::DomainRuntime;
use super::totem_runtime::TotemRuntime;


/// Process-local container for running domains.
pub struct NodeRuntime {
    /// All running domains keyed by their intrinsic DomainId.
    domains: HashMap<DomainId, DomainRuntime>,

    /// Process-local totem presence (runtime fact).
    totem: TotemRuntime,
}


impl NodeRuntime {
    // ------------------------------------------------------------
    // Construction
    // ------------------------------------------------------------

    /// Create a new, empty NodeRuntime.
    ///
    /// This does not load domains, start services, or perform I/O.
    pub fn new() -> Self {
        Self {
            domains: HashMap::new(),
            totem: TotemRuntime::new(),
        }
    }

    // ------------------------------------------------------------
    // Domain lifecycle
    // ------------------------------------------------------------

    /// Insert a domain and create its running runtime.
    ///
    /// - The Domain must already be identity-anchored (have a DomainId).
    /// - No validation, persistence, or execution occurs here.
    ///
    /// Returns a reference to the newly created DomainRuntime.
    pub fn insert_domain(&mut self, domain: Domain) -> &DomainRuntime {
        let id = domain.id.clone();

        let runtime = DomainRuntime::new(domain);

        self.domains.insert(id.clone(), runtime);

        self.domains
            .get(&id)
            .expect("domain runtime must exist immediately after insertion")
    }

    /// Remove a running domain by id.
    ///
    /// This drops the DomainRuntime and all associated runtime state.
    pub fn remove_domain(&mut self, id: &DomainId) -> Option<DomainRuntime> {
        self.domains.remove(id)
    }

    // ------------------------------------------------------------
    // Lookup
    // ------------------------------------------------------------

    /// Get a running domain by id.
    pub fn get_domain(&self, id: &DomainId) -> Option<&DomainRuntime> {
        self.domains.get(id)
    }

    /// Get a mutable reference to a running domain.
    ///
    /// This is intentionally explicit to avoid accidental mutation.
    pub fn get_domain_mut(&mut self, id: &DomainId) -> Option<&mut DomainRuntime> {
        self.domains.get_mut(id)
    }

    /// Check whether a domain is currently running.
    pub fn contains_domain(&self, id: &DomainId) -> bool {
        self.domains.contains_key(id)
    }

    // ------------------------------------------------------------
    // Introspection
    // ------------------------------------------------------------

    /// Iterate over all running domains.
    pub fn iter_domains(&self) -> impl Iterator<Item = &DomainRuntime> {
        self.domains.values()
    }

    /// Iterate over all running domain ids.
    pub fn iter_domain_ids(&self) -> impl Iterator<Item = &DomainId> {
        self.domains.keys()
    }

    /// Number of running domains.
    pub fn domain_count(&self) -> usize {
        self.domains.len()
    }

    // ------------------------------------------------------------
    // Totem presence (node-scoped)
    // ------------------------------------------------------------

    /// Get a reference to the node's totem runtime presence.
    pub fn totem(&self) -> &TotemRuntime {
        &self.totem
    }

    /// Get a mutable reference to the node's totem runtime presence.
    ///
    /// This is intentionally explicit to avoid accidental mutation.
    pub fn totem_mut(&mut self) -> &mut TotemRuntime {
        &mut self.totem
    }
}

