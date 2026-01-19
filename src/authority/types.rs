// ============================================================
// FILE: src/authority/types.rs
// ============================================================

use core::fmt;

// -----------------------------
// Fundamental references
// -----------------------------

#[derive(Clone, Debug, PartialEq, Eq, Hash)]
pub struct ActorRef {
    // Nullus anchor (ID only). Keep opaque.
    pub id: String,
}

#[derive(Clone, Debug, PartialEq, Eq, Hash)] 
pub struct DomainRef {
    // Terra anchor (scope root).
    pub id: String,
}

#[derive(Clone, Debug, PartialEq, Eq, Hash)]
pub struct Scope {
    // Scoped authority target (freeze scope, commit scope).
    // Keep as stable string for now; refine later.
    pub key: String,
}

// -----------------------------
// Intent (Numen / Lima carrier)
// -----------------------------

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Intent {
    // Canonicalized intent bytes hash (NOT Lima text).
    // Caller supplies canonicalization result; kernel trusts only hash.
    pub canonical_hash: [u8; 32],
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PolicyRef {
    // Reference to policy decision provenance (Phase-2 MinSet).
    pub id: String,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ProvenanceRef {
    // Reference to CI provenance / SAT chain / external attestations.
    pub id: String,
}

// -----------------------------
// Freeze
// -----------------------------

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum FreezeOp {
    Freeze,        // domain.freeze.v1
    Unfreeze,      // domain.unfreeze.v1
    BreakGlass,    // domain.freeze.override.v1
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct FreezeIntent {
    pub op: FreezeOp,
    pub domain: DomainRef,
    pub scope: Scope,
    // TTL is data only. No clocks here; downstream evaluators may interpret later.
    pub ttl_hint_seconds: Option<u64>,
    pub reason: String, // human-readable reason, will be redacted in receipt by default
}

// -----------------------------
// Commit
// -----------------------------

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CommitIntent {
    pub domain: DomainRef,
    pub scope: Scope,
    pub intent: Intent,
    // The thing being committed (opaque handle); could become artifact ref later.
    pub target: String,
}

// -----------------------------
// Requests / Outcomes
// -----------------------------

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum AuthorityRequest {
    Freeze {
        // Corporeal actor performing the act
        actor: ActorRef,
        // Numen/Lima expressed request (canonicalized)
        intent: FreezeIntent,
        // Mandatory provenance references
        policy: PolicyRef,
        provenance: ProvenanceRef,
    },
    Commit {
        actor: ActorRef,
        intent: CommitIntent,
        policy: PolicyRef,
        provenance: ProvenanceRef,
    },
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum AuthorityOutcome {
    Allowed {
        receipt: ReceiptRef,
        // Optional additional sealed data (commit id, freeze state id) without leaking power
        sealed: SealedOutcomeData,
    },
    Denied {
        receipt: ReceiptRef,
        reason: super::errors::DenyReason,
    },
    Error(super::errors::AuthorityError),
}

// Keep outcome data sealed and minimal; no internals.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum SealedOutcomeData {
    None,
    FreezeStateRef(String),
    CommitRef(String),
}

// -----------------------------
// Receipt (minimal public shape)
// -----------------------------

#[derive(Clone, Debug, PartialEq, Eq, Hash)]
pub struct ReceiptRef {
    pub id: String,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Receipt {
    pub id: ReceiptRef,
    pub kind: ReceiptKind,
    pub actor: ActorRef,
    pub domain: DomainRef,
    pub scope: Scope,
    pub outcome: ReceiptOutcome,
    // References only — never embed sensitive data by default.
    pub policy: PolicyRef,
    pub provenance: ProvenanceRef,
    // Signature placeholder; implement later (Phase 3.2+)
    pub signature: Option<Vec<u8>>,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum ReceiptKind {
    CiCallV1, // Aligns to ci.call.v1 concept; keep name stable
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum ReceiptOutcome {
    Allowed,
    Denied { code: String },
    Error { code: String },
}

impl fmt::Display for ReceiptRef {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        f.write_str(&self.id)
    }
}


