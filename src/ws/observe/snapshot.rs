// src/ws/observe/snapshot.rs
use crate::context::RuntimeContext;
use crate::ws::observe::messages::*;
use crate::id::DomainId;

// TEMP: until observe becomes domain-scoped
// choose a single implicit domain if one exists
fn default_domain(_runtime: &RuntimeContext) -> Option<DomainId> {
    // You can refine this later (active domain, subscription, etc.)
    None
}

pub fn build_snapshot(
    runtime: &RuntimeContext,
) -> ObserveSnapshot {
    let domain = default_domain(runtime);

    let chat = domain.map(|domain_id| {
        runtime.chat().get_or_create(domain_id)
    });

    ObserveSnapshot {
        tick: 0, // later: derive from chat.tick or observation clock

        node: NodeView {
            name: "dis-core",
            version: "0.1.0",
        },

        domain: None, // domain identity comes later

        spine: SpineView {
            nullus: true,
            aether: false,
            terra: false,
            numen: false,
            lima: false,
            corporeal: false,
        },

        capabilities: CapabilityView {
            can_bind_identity: false,
            can_claim_sovereignty: false,
        },

        chat, // 👈 HERE IT IS
    }
}
