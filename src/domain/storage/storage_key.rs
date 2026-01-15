use crate::domain::lattice_axis::LatticeAxis;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct StorageKey {
    /// Lattice axis this storage is bound to
    pub axis: LatticeAxis,

    /// Opaque material discriminator (hash, salt, etc.)
    pub material: [u8; 32],
}
