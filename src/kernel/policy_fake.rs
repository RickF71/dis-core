use crate::kernel::decision::{Decision, ReasonCode};
use crate::kernel::identity::DomainId;
use crate::policy::policy_ref::PolicyRef;
use crate::spine::tick::DisTick;
use super::decision::{Decision, ReasonCode};
use dis_core::domain::domain_id::DomainId;

// Stub types for PolicyRef and DisTick (remove or replace when available)
#[derive(Debug, Clone)]
pub struct PolicyRef;
#[derive(Debug, Clone, Copy)]
pub struct DisTick;

pub fn fake_policy_decision(
    domain: DomainId,
    tick: DisTick,
    allow: bool,
) -> Decision {
    if allow {
        Decision {
            allow: true,
            reason: ReasonCode::Allowed,
            policy_ref: PolicyRef::fake_allow(),
            domain,
            tick,
        }
    } else {
        Decision::deny_by_default(domain, tick)
    }
}
