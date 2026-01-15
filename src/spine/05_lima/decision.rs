// src/spine/05_lima/decision.rs
use crate::spine::lima::reasons::DenyReason;

#[derive(Debug, Clone)]
pub enum Decision {
    Allow,
    Deny { reason: DenyReason },
}

impl Decision {
    pub fn deny(reason: DenyReason) -> Self {
        Decision::Deny { reason }
    }
}
