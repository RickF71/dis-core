/// src/context/observation.rs
/// A stable frame for observing domains.
/// Immutable once created.
///
/// Represents a 6-D observation coordinate.
/// This is NOT a clock and carries no authority.


#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ObservationFrame {
    pub sequence: u64,
}

impl ObservationFrame {
    pub fn new(sequence: u64) -> Self {
        Self { sequence }
    }
}


/// Process-local observation coordinator.
///
/// Mutable, internal, and NEVER serialized.
#[derive(Debug, Default)]
pub struct ObservationState {
    sequence: u64,
}

impl ObservationState {
    pub fn new() -> Self {
        Self { sequence: 0 }
    }

    /// Advance and produce a new observation frame
    pub fn next_frame(&mut self) -> ObservationFrame {
        self.sequence += 1;
        ObservationFrame::new(self.sequence)
    }

    /// Read current frame without advancing
    pub fn current_frame(&self) -> ObservationFrame {
        ObservationFrame::new(self.sequence)
    }
}
