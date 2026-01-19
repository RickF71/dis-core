// ============================================================
// FILE: src/authority/errors.rs
// ============================================================

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum AuthorityError {
    // Structural violations (hard abort)
    MissingIdentityBinding,
    NonBypassViolation,
    InvalidScope,
    InvalidPolicyRef,
    InvalidProvenanceRef,
    KernelMisconfiguration,
    InternalInvariantFailed(&'static str),

    // Phase 3.7 — lineage validation
    ParentNotFound,
    ParentDomainMismatch,
    ParentCycleDetected,
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum DenyReason {
    // Policy/Freeze denies (soft deny, still receipted)
    FreezeActive { scope: String }, // deny:freeze:<scope>
    PolicyDenied { policy_ref: String },
    ActorNotAuthorized,
    ScopeNotAdopted,
    Other { code: String },
}

