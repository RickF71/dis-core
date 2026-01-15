// src/ws/totem/messages.rs
use serde::Serialize;
use serde::Deserialize;

// ================================
// Totem presence & state messages
// Published by dis-core to Finagler
// ================================

#[derive(Serialize)]
#[serde(tag = "type", content = "payload")]
pub enum TotemMessage {
    #[serde(rename = "totem.status.v1")]
    Status(TotemStatus),

    #[serde(rename = "totem.offline.v1")]
    Offline(TotemOffline),

    #[serde(rename = "totem.capabilities.v1")]
    Capabilities(TotemCapabilities),

    #[serde(rename = "totem.health.v1")]
    Health(TotemHealth),
}

// -------------------------------
// Core presence signal
// -------------------------------

#[derive(Serialize)]
pub struct TotemStatus {
    pub present: bool,

    /// RFC3339 timestamp, set only when present = true
    pub since: Option<String>,

    /// Milliseconds since last heartbeat
    pub last_seen_ms: Option<u64>,
}

// -------------------------------
// Explicit absence event
// -------------------------------

#[derive(Serialize)]
pub struct TotemOffline {
    pub reason: String, // e.g. "heartbeat_timeout", "process_exit"
}

// -------------------------------
// Capability surface
// -------------------------------

#[derive(Serialize)]
pub struct TotemCapabilities {
    pub capabilities: Vec<String>,
}

// -------------------------------
// Health (not logs)
// -------------------------------

#[derive(Serialize)]
pub struct TotemHealth {
    pub state: String, // ok | degraded | stalled | failed
    pub latency_ms: Option<u64>,
    pub warnings: Vec<String>,
}


#[derive(Deserialize)]
#[serde(tag = "type", content = "payload")]
pub enum TotemInbound {
    #[serde(rename = "totem.heartbeat.v1")]
    Heartbeat(TotemHeartbeat),
}

#[derive(Deserialize)]
pub struct TotemHeartbeat {}
