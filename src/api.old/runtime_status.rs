// src/api/runtime_status.rs
use serde::Serialize;
use std::time::Duration;

use crate::runtime::node_runtime::NodeRuntime;

#[derive(Serialize)]
pub struct RuntimeStatus {
    pub totem: RuntimePresence,
    pub node: RuntimePresence,
    pub domains: DomainsPresence,
}

#[derive(Serialize)]
pub struct RuntimePresence {
    pub present: bool,
    pub last_seen_ms: Option<u64>,
}

#[derive(Serialize)]
pub struct DomainsPresence {
    pub running: usize,
}

impl RuntimeStatus {
    pub fn from_node(
        node: &NodeRuntime,
        timeout: Duration,
    ) -> Self {
        let totem = node.totem();

        Self {
            totem: RuntimePresence {
                present: totem.is_present(timeout),
                last_seen_ms: totem.last_seen_ms(),
            },
            node: RuntimePresence {
                present: true,
                last_seen_ms: None,
            },
            domains: DomainsPresence {
                running: node.domain_count(),
            },
        }
    }
}
