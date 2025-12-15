use crate::spine::Layer6;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Capsule<T> {
    pub payload: T,
    pub from: Layer6,
    pub to: Layer6,
    pub tick: u64, // traversal depth (one per edge)
}

impl<T> Capsule<T> {
    /// Create the initial capsule (Nullus admission)
    pub fn new(payload: T, from: Layer6, to: Layer6, tick: u64) -> Self {
        Self { payload, from, to, tick }
    }

    /// Advance exactly one spine edge.
    ///
    /// This is the ONLY place ticks are incremented.
    pub fn tick_to(mut self, to: Layer6) -> Self {
        self.from = self.to;
        self.to = to;
        self.tick += 1;
        self
    }

    /// Change routing target without advancing time.
    /// Use only for metadata correction.
    pub fn with_layer(mut self, to: Layer6) -> Self {
        self.to = to;
        self
    }
}
