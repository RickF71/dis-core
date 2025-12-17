use dis_core::domain::domain_id::DomainId;

// Stub types for PolicyRef and DisTick (remove or replace when available)
#[derive(Debug, Clone)]
pub struct PolicyRef;
#[derive(Debug, Clone, Copy)]
pub struct DisTick;

#[derive(Debug, Clone)]
pub struct Decision {
    pub allow: bool,
    pub reason: ReasonCode,
    // pub policy_ref: PolicyRef, // Remove until PolicyRef exists
    pub domain: DomainId,
    // pub tick: DisTick, // Remove until DisTick exists
}

impl Decision {
    /// Canonical deny-by-default constructor (MinSet-5)
    pub fn deny_by_default(
        domain: DomainId,
        // tick: DisTick,
    ) -> Self {
        Self {
            allow: false,
            reason: ReasonCode::DenyByDefault,
            // policy_ref: PolicyRef::system_default(),
            domain,
            // tick,
        }
    }
}

#[derive(Debug, Clone)]
pub enum ReasonCode {
    Allowed,
    DenyFreeze { scope: String },
    DenyPolicy { rule: String },
    DenyInvalidIdentity,
    DenyByDefault,
}
