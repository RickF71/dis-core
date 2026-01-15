// src/taiji/emitter.rs

use std::sync::{Arc, RwLock};

use crate::taiji::snapshot::Snapshot6D;

/// Owns the latest snapshot and decides when to emit.
#[derive(Clone)]
pub struct SnapshotEmitter {
    latest: Arc<RwLock<Option<Snapshot6D>>>,
}

impl SnapshotEmitter {
    pub fn new() -> Self {
        Self {
            latest: Arc::new(RwLock::new(None)),
        }
    }

    /// Set a new snapshot.
    /// Returns true if it replaced a previous snapshot.
    pub fn update(&self, snapshot: Snapshot6D) -> bool {
        let mut guard = self.latest.write().unwrap();
        let replaced = guard.is_some();
        *guard = Some(snapshot);
        replaced
    }

    /// Get the current snapshot (if any).
    pub fn current(&self) -> Option<Snapshot6D> {
        self.latest.read().unwrap().clone()
    }
}
