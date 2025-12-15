
#[derive(Debug, Clone)]
pub struct DisClock {
    pub dis_tick: u64,
    pub phase: SpinePhase,
}

impl DisClock {
    pub fn new() -> Self {
        Self {
            dis_tick: 1,
            phase: SpinePhase::Nullus,
        }
    }

    pub fn commit_allowed(&self) -> bool {
        matches!(self.phase, SpinePhase::Corporeal)
    }
}
use serde::{Serialize, Deserialize};

/// One observable unit of DIS time.
/// Advances only after Corporeal completes.
pub type DisTick = u64;

/// Internal deterministic sub-tick.
/// Exactly one per spine layer, in order.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]

#[repr(u8)]
pub enum SpinePhase {
    Nullus    = 1,
    Aether    = 2,
    Terra     = 3,
    Numen     = 4,
    Lima      = 5,
    Corporeal = 6,
}


impl SpinePhase {
    pub const ALL: [SpinePhase; 6] = [
        SpinePhase::Nullus,
        SpinePhase::Aether,
        SpinePhase::Terra,
        SpinePhase::Numen,
        SpinePhase::Lima,
        SpinePhase::Corporeal,
    ];
}
