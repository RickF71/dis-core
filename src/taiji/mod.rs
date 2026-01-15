// src/taiji/mod.rs

pub mod taiji;        // identity + wiring
pub mod clock;        // time / ticks
pub mod errors;       // taiji-specific errors

pub mod snapshot;     // Snapshot6D + view types (canonical)
pub mod emitter;      // 👈 NEW: snapshot emission state
pub mod projection;   // future: mapping snapshot → UI / lenses

pub use taiji::Taiji;
pub use emitter::SnapshotEmitter;
