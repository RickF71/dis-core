use crate::id::DomainId;
use crate::domain::lattice::DomainLattice;

#[derive(Debug, Clone)]
pub struct ReflectiveSurface {
    /// Optional parent domain (lineage)
    pub parent_domain: Option<DomainId>,

    /// Lattice coordinate of the last acknowledgment by the parent
    pub last_ack_lattice: Option<DomainLattice>,

    /// Whether this domain is recognized by its parent
    pub recognized: bool,

    /// Whether this domain is currently frozen
    pub frozen: bool,
}

impl ReflectiveSurface {
    /// Initial state for an unknown domain
    pub fn unknown() -> Self {
        Self {
            parent_domain: None,
            last_ack_lattice: None,
            recognized: false,
            frozen: false,
        }
    }
}
