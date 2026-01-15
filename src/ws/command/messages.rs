// src/ws/command/messages.rs
use serde::{Deserialize, Serialize};

#[derive(Deserialize)]
#[serde(tag = "type", content = "payload")]
pub enum CommandMessage {
    #[serde(rename = "identity.bind.request.v1")]
    IdentityBind(IdentityBindRequest),

    #[serde(rename = "sovereignty.claim.request.v1")]
    SovereigntyClaim(SovereigntyClaimRequest),
}

#[derive(Deserialize)]
pub struct IdentityBindRequest {
    pub proof: String,
}

#[derive(Deserialize)]
pub struct SovereigntyClaimRequest {
    pub declaration: String,
}

#[derive(Serialize)]
#[serde(tag = "type", content = "payload")]
pub enum CommandResponse {
    #[serde(rename = "command.accepted.v1")]
    Accepted { intent_id: String },

    #[serde(rename = "command.denied.v1")]
    Denied { intent_id: String, reason: String },
}
