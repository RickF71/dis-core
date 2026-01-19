// src/runtime/record.rs

use crate::store::{Store, artifact::Artifact};
use crate::runtime::commit::CommitKind;

#[derive(Debug)]
pub enum RecordError {
    Store(std::io::Error),
    Invariant(&'static str),
}

impl From<std::io::Error> for RecordError {
    fn from(e: std::io::Error) -> Self {
        RecordError::Store(e)
    }
}

/// Materialize an already-authorized commit into durable memory.
/// This function does NOT decide policy.
pub async fn record_commit(
    store: &Store,
    _commit: CommitKind,
    artifact: Artifact,
) -> Result<(), RecordError> {
    // Invariant 1: artifact kind must match commit kind
    // (you can relax this later, but enforce *something* now)
    if artifact.kind.is_empty() {
        return Err(RecordError::Invariant("artifact.kind must be set"));
    }

    // Invariant 2: Coord6 must be present (already enforced by type)
    // Invariant 3: record is append-only (enforced by Store)

    store.record(&artifact).await?;
    Ok(())
}
