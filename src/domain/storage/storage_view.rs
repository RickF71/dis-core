// src/spine/storage_view.rs

use std::marker::PhantomData;

use crate::spine::Layer6;
use crate::spine::payload::{PayloadId, PayloadShapeSchema, ViewContract};
use crate::domain::storage::storage_key::StorageKey;

/// A backend-agnostic locator for where bytes live.
/// File path, DB row, object store key, etc.
/// Opaque by design.
#[derive(Debug, Clone)]
pub struct Locator(pub String);

/// A typed storage lens:
/// - bound to a payload shape schema S
/// - bound to a specific layer view V (one of S's six views)
/// - gated by a StorageKey for V::LAYER
#[derive(Debug, Clone)]
pub struct StorageView<S, V>
where
    S: PayloadShapeSchema,
    V: ViewContract,
{
    pub payload: PayloadId,
    pub domain: Layer6,     // redundant but convenient; must match S::DOMAIN
    pub layer: Layer6,      // must match V::LAYER
    pub key: StorageKey,    // must match layer
    pub locator: Locator,   // opaque pointer to implementation
    _shape: PhantomData<S>,
    _view: PhantomData<V>,
}

impl<S, V> StorageView<S, V>
where
    S: PayloadShapeSchema,
    V: ViewContract,
{
    /// Construct a StorageView with invariant checks.
    /// (Runtime check is fine here; type system already does the heavy lifting.)
    pub fn new(payload: PayloadId, key: StorageKey, locator: Locator) -> Self {
        // Lock: key must match view layer
        assert_eq!(key.layer, V::LAYER, "deny:storage:key_layer_mismatch");

        Self {
            payload,
            domain: S::DOMAIN,
            layer: V::LAYER,
            key,
            locator,
            _shape: PhantomData,
            _view: PhantomData,
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct ReceiptId(pub [u8; 32]);
