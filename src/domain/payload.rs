use crate::domain::lattice_axis::LatticeAxis;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct PayloadId(pub [u8; 32]);

/// Reference to a payload within a domain and lattice context.
/// This is a coordinate, not a type-level projection.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct PayloadRef {
    pub axis: LatticeAxis,
    pub id: PayloadId,
}
