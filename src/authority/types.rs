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
    Freeze,     // domain.freeze.v1
    Unfreeze,   // domain.unfreeze.v1
    BreakGlass, // domain.freeze.override.v1
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
        actor: ActorRef,
        intent: FreezeIntent,
        policy: PolicyRef,
        provenance: ProvenanceRef,

        // Phase 3.5: optional lineage input
        parent: Option<ReceiptRef>,
    },
    Commit {
        actor: ActorRef,
        intent: CommitIntent,
        policy: PolicyRef,
        provenance: ProvenanceRef,

        // Phase 3.5: optional lineage input
        parent: Option<ReceiptRef>,
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

    // Optional lineage pointer (Phase 3.5)
    pub parent: Option<ReceiptRef>,

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

// ============================================================
// Phase 3.3 — Receipt sealing boundary (types)
// Additive: introduces Envelope + Seal without changing kernel logic yet
// ============================================================

/// A signature/seal applied to a Receipt *outside* the authority kernel.
///
/// IMPORTANT:
/// - Authority kernel does NOT create these.
/// - Transport does NOT interpret them.
/// - They are opaque bytes + metadata to support future verification.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ReceiptSeal {
    /// e.g. "ed25519", "p256", "domain.sig.v1"
    pub scheme: String,

    /// key identifier or reference (NOT the key itself)
    pub key_id: String,

    /// signature bytes (opaque)
    pub sig: Vec<u8>,
}

/// Optional witness references to attach to an envelope later.
/// (CI runs, SAT chains, external attestations, etc.)
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct WitnessRef {
    pub id: String,
}

/// ReceiptEnvelope is the *signed/attested* form of a receipt.
/// The authority kernel produces `Receipt` (core) only.
/// External layers may wrap it with seals and witnesses.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ReceiptEnvelope {
    pub core: Receipt,
    pub seals: Vec<ReceiptSeal>,
    pub witnesses: Vec<WitnessRef>,
}

/// Canonical, deterministic signable form.
///
/// No crypto yet. No hashing yet.
/// This is purely “what bytes would be signed”.
///
/// Rules:
/// - Must be stable.
/// - Must not include anything that can vary by transport.
/// - Must not include `seals`/`witnesses` (those belong to the envelope).
impl Receipt {
    pub fn signable_string_v1(&self) -> String {
        let outcome_str = match &self.outcome {
            ReceiptOutcome::Allowed => "allowed".to_string(),
            ReceiptOutcome::Denied { code } => format!("denied:{}", code),
            ReceiptOutcome::Error { code } => format!("error:{}", code),
        };

        let parent_str = match &self.parent {
            None => "none".to_string(),
            Some(p) => p.id.clone(),
        };

        // NOTE: signatures wrap the receipt; the receipt does not contain them.
        format!(
            "receipt.signable.v1|id={}|parent={}|kind={:?}|actor={}|domain={}|scope={}|outcome={}|policy={}|prov={}",
            self.id.id,
            parent_str,
            self.kind,
            self.actor.id,
            self.domain.id,
            self.scope.key,
            outcome_str,
            self.policy.id,
            self.provenance.id,
        )
    }
}

// Add this BELOW your existing signable_string_v1() impl block

impl Receipt {
    /// Phase 3.8: lineage-binding signable form (includes parent).
    /// This is intentionally additive (v1 remains stable).
    pub fn signable_string_v2(&self) -> String {
        let outcome_str = match &self.outcome {
            ReceiptOutcome::Allowed => "allowed".to_string(),
            ReceiptOutcome::Denied { code } => format!("denied:{}", code),
            ReceiptOutcome::Error { code } => format!("error:{}", code),
        };

        let parent_str = match &self.parent {
            Some(p) => p.id.as_str(),
            None => "",
        };

        format!(
            "receipt.signable.v2|id={}|parent={}|kind={:?}|actor={}|domain={}|scope={}|outcome={}|policy={}|prov={}",
            self.id.id,
            parent_str,
            self.kind,
            self.actor.id,
            self.domain.id,
            self.scope.key,
            outcome_str,
            self.policy.id,
            self.provenance.id,
        )
    }
}
