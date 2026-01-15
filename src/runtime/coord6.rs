use crate::runtime::commit::CommitKind;

#[derive(Debug, Clone, Copy, serde::Serialize)]
pub struct Coord6 {
    pub n: u64,
    pub a: u64,
    pub t: u64,
    pub nu: u64,
    pub l: u64,
    pub c: u64,
}

impl Coord6 {
    pub fn zero() -> Self {
        Self { n: 0, a: 0, t: 0, nu: 0, l: 0, c: 0 }
    }

    pub fn advance(&mut self, kind: CommitKind) {
        match kind {
            CommitKind::IdentityChange     => self.n += 1,
            CommitKind::MemoryAppend       => self.a += 1,
            CommitKind::StructuralChange   => self.t += 1,
            CommitKind::MeaningBind        => self.nu += 1,
            CommitKind::InterpretationBind => self.l += 1,
            CommitKind::Externalize        => self.c += 1,
        }
    }
}
