// ============================================================
// Phase 3 — Authority Kernel Skeleton (dis-core-rust)
// Goal: define where authority lives, how exercised, and how receipts mint
// Constraints: Phase-2 quarantine (no clocks, no runtime leakage, no transport leakage)
// ============================================================
//
// Recommended crate layout additions:
//
// src/
//   authority/
//     mod.rs
//     prelude.rs
//     types.rs
//     errors.rs
//     seal.rs
//     gate.rs
//     freeze.rs
//     commit.rs
//     receipt.rs
//   lib.rs  (re-export only what is intended)
// tests/
//   authority_freeze.rs
//   authority_commit.rs
//
// NOTE: This is a skeleton: compile-first, wire-later.
// No storage implementation required here; define traits only.
//
// ============================================================
// FILE: src/authority/mod.rs
// ============================================================

pub mod prelude;
pub mod types;
pub mod errors;

mod seal;
mod gate;
mod freeze;
mod commit;
mod receipt;

// Public kernel API: keep it small and boring.
// Everything else is internal.
pub use gate::{AuthorityKernel, AuthorityKernelConfig};
pub use types::{
    AuthorityRequest,
    AuthorityOutcome,
    Receipt,
    ReceiptRef,
    ActorRef,
    DomainRef,
    Scope,
    Intent,
    CommitIntent,
    FreezeIntent,
    FreezeOp,
    PolicyRef,
    ProvenanceRef,
};
pub use errors::{AuthorityError, DenyReason};




// ============================================================
// FILE: tests/authority_freeze.rs (skeleton)
// ============================================================
//
// - Provide fake reader/writer
// - Assert receipts always append
// - Assert freeze op returns Allowed + FreezeStateRef
//
// ============================================================
// FILE: tests/authority_commit.rs (skeleton)
// ============================================================
//
// - When frozen => Denied with deny:freeze:<scope> in receipt outcome
// - When not frozen => Allowed + CommitRef
//
// ============================================================
// Phase Handoff Summary Template (JSON)
// ============================================================
//
// {
//   "phase": 3,
//   "name": "Authority & Commitment",
//   "status": "skeleton_proposed",
//   "new_modules": [
//     "src/authority/mod.rs",
//     "src/authority/types.rs",
//     "src/authority/errors.rs",
//     "src/authority/seal.rs",
//     "src/authority/receipt.rs",
//     "src/authority/gate.rs"
//   ],
//   "public_surface": [
//     "AuthorityKernel::apply(AuthorityRequest) -> AuthorityOutcome",
//     "AuthorityRequest::{Freeze,Commit}",
//     "ReceiptRef + minimal Receipt shape"
//   ],
//   "invariants": [
//     "No clocks",
//     "No transport imports",
//     "No ontology mutation",
//     "All authority produces receipt",
//     "Freeze gates commit"
//   ],
//   "next_steps": [
//     "Replace uuid_stub() with injected sequence/entropy trait (still no clocks)",
//     "Write fake in-memory implementations for FreezeState/Commit/Receipt to compile tests",
//     "Wire policy_ref/provenance_ref validation to Phase-2 policy decision presence"
//   ]
// }


#[cfg(test)]
mod tests;


