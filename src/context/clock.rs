// src/runtime/clock.rs
#[derive(Debug, Clone)]
pub struct RuntimeClock {
    tick: u64,
}

impl RuntimeClock {
    pub fn new() -> Self {
        Self { tick: 1 }
    }

    /// Advance the observation clock.
    /// This does NOT mutate domain state.
    pub fn advance(&mut self) {
        self.tick += 1;
    }

    /// Current observation tick.
    pub fn tick(&self) -> u64 {
        self.tick
    }
}
