use crate::domain::domain_id::DomainId;
use crate::spine::clock::DisTick;

#[derive(Debug, Clone)]
pub struct ReflectiveSurface {
    pub parent_domain: Option<DomainId>,
    pub last_ack_tick: Option<DisTick>,
    pub recognized: bool,
    pub frozen: bool,
}

impl ReflectiveSurface {
    pub fn unknown() -> Self {
        Self {
            parent_domain: None,
            last_ack_tick: None,
            recognized: false,
            frozen: false,
        }
    }
}
