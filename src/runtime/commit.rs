// src/runtime/commit.rs
#[derive(Debug, Clone, Copy, PartialEq, Eq, serde::Serialize)]
pub enum CommitKind {
    IdentityChange,
    MemoryAppend,
    StructuralChange,
    MeaningBind,
    InterpretationBind,
    Externalize,
}
