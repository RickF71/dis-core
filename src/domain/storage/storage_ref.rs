use crate::domain::lattice_axis::LatticeAxis;
use crate::domain::payload::PayloadId;
use crate::domain::storage::storage_key::StorageKey;

/// A backend-agnostic locator for where bytes live.
/// File path, DB row, object store key, etc.
#[derive(Debug, Clone)]
pub struct Locator(pub String);

/// A storage reference:
/// - scoped to a payload
/// - scoped to a lattice axis
/// - guarded by a storage key
/// - pointing to opaque backend storage
#[derive(Debug, Clone)]
pub struct StorageRef {
    pub payload: PayloadId,
    pub axis: LatticeAxis,
    pub key: StorageKey,
    pub locator: Locator,
}

impl StorageRef {
    /// Construct a storage reference with invariant checks.
    pub fn new(
        payload: PayloadId,
        axis: LatticeAxis,
        key: StorageKey,
        locator: Locator,
    ) -> Self {
        // Axis consistency check
        assert_eq!(
            axis, key.axis,
            "deny:storage:axis_key_mismatch"
        );

        Self {
            payload,
            axis,
            key,
            locator,
        }
    }
}
