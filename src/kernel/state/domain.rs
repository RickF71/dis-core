
use std::collections::HashSet;
use dis_core::domain::domain_id::DomainId;

#[derive(Clone)]
pub struct DomainState {
    frozen: HashSet<DomainId>,
}

impl DomainState {
    pub fn new() -> Self {
        Self {
            frozen: HashSet::new(),
        }
    }

    /// TEMP helper for testing / bootstrap
    pub fn freeze_domain(&mut self, domain: DomainId) {
        self.frozen.insert(domain);
    }

    pub fn is_frozen(&self, domain: &DomainId) -> bool {
        self.frozen.contains(domain)
    }

}
