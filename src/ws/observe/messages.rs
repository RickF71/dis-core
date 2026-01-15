// src/ws/observe/messages.rs
use serde::Serialize;

use crate::chat::ChatRoom;

#[derive(Serialize)]
#[serde(tag = "type", content = "payload")]
pub enum ObserveMessage {
    #[serde(rename = "observe.snapshot.v1")]
    Snapshot(ObserveSnapshot),
}

#[derive(Serialize)]
pub struct ObserveSnapshot {
    pub tick: u64,
    pub node: NodeView,
    pub domain: Option<DomainView>,
    pub spine: SpineView,
    pub capabilities: CapabilityView,

    // 👇 THIS IS WHY ChatRoom IS IMPORTED
    pub chat: Option<ChatRoom>,
}

#[derive(Serialize)]
pub struct NodeView {
    pub name: &'static str,
    pub version: &'static str,
}

#[derive(Serialize)]
pub struct DomainView {
    pub id: String,
    pub layer: &'static str,
    pub authority_present: bool,
}

#[derive(Serialize)]
pub struct SpineView {
    pub nullus: bool,
    pub aether: bool,
    pub terra: bool,
    pub numen: bool,
    pub lima: bool,
    pub corporeal: bool,
}

#[derive(Serialize)]
pub struct CapabilityView {
    pub can_bind_identity: bool,
    pub can_claim_sovereignty: bool,
}
