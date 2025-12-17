use dis_core::domain::domain_id::DomainId;
use dis_core::spine::capsule::Capsule;

use super::decision::Decision;
use super::identity::IdentityRef;
use super::state::DomainState;

/// Opaque reference to a decision/witness artifact.
/// (Later: this becomes a seat-stored receipt id, or a witness event id.)
pub type DecisionRef = String;

/// Result of attempting a commit at the kernel boundary.
#[derive(Debug)]
pub enum CommitResult {
    Applied { decision_ref: DecisionRef },
    Denied  { decision_ref: DecisionRef },
}

/// CommitBoundary is the sole authority gate for state changes.
/// All domain mutations MUST pass through this interface.
pub trait CommitBoundary {
    fn commit(
        &mut self,
        domain: &DomainId,
        capsule: Capsule<()>,
        identity: &IdentityRef,
        decision: Decision,
    ) -> CommitResult;
}

/// KernelCommitter is the concrete authority that enforces commit rules.
/// At MinSet-5, this is deny-by-default (receipts/witnessing happen outside kernel state).
#[derive(Clone)]
pub struct KernelCommitter {
    state: DomainState,
}

impl KernelCommitter {
    pub fn new(state: DomainState) -> Self {
        Self { state }
    }

    fn decision_ref(
        domain: &DomainId,
        identity: &IdentityRef,
        decision: &Decision,
    ) -> DecisionRef {
        format!("dec:{:?}:{:?}:{:?}", domain, identity, decision)
    }

}

impl CommitBoundary for KernelCommitter {
    fn commit(
        &mut self,
        domain: &DomainId,
        _capsule: Capsule<()>,
        identity: &IdentityRef,
        decision: Decision,
    ) -> CommitResult {
        // MinSet-5: deny by default (no state mutation yet).
        let decision_ref = Self::decision_ref(domain, identity, &decision);

        // NOTE: when you reintroduce witnessing, emit a WitnessEvent here,
        // and let seat-owned storage materialize receipts.
        CommitResult::Denied { decision_ref }
    }
}
