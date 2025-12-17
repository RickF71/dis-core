// src/kernel/mod.rs

use dis_core::domain::domain_id::DomainId;
use dis_core::spine::capsule::Capsule;

mod commit;
mod decision;
mod identity;
mod state;

pub use commit::{CommitBoundary, CommitResult, KernelCommitter};
pub use decision::Decision;
pub use identity::IdentityRef;
pub use state::DomainState;

#[derive(Clone)]
pub struct Kernel {
    state: DomainState,
}

impl Kernel {
    pub fn new() -> Self {
        Self {
            state: DomainState::new(),
            // receipts: Arc::new(Mutex::new(Vec::new())),
        }
    }

    /// Single public commit entry point (MinSet-5 choke point).
    pub fn commit(
        &self,
        domain: &DomainId,
        capsule: Capsule<()>,
        identity: &IdentityRef,
        decision: Decision,
    ) -> CommitResult {
        let mut committer = KernelCommitter::new(self.state.clone());
        let result = committer.commit(domain, capsule, identity, decision);



        result
    }


}
