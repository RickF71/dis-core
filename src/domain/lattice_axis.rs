use serde::{Serialize, Deserialize};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[repr(u8)]
pub enum LatticeAxis {
    Nullus    = 1,
    Aether    = 2,
    Terra     = 3,
    Numen     = 4,
    Lima      = 5,
    Corporeal = 6,
}

impl LatticeAxis {
    pub const ALL: [LatticeAxis; 6] = [
        LatticeAxis::Nullus,
        LatticeAxis::Aether,
        LatticeAxis::Terra,
        LatticeAxis::Numen,
        LatticeAxis::Lima,
        LatticeAxis::Corporeal,
    ];

    pub fn index(self) -> usize {
        (self as u8 - 1) as usize
    }
}
