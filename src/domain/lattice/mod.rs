// src/domain/lattice/mod.rs

use serde::{Serialize, Deserialize};
use crate::domain::lattice_axis::LatticeAxis;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub struct DomainLattice {
    pub axis: LatticeAxis,
}
