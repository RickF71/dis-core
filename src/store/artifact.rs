use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Coord6 {
    pub n: u64,
    pub a: u64,
    pub t: u64,
    pub nu: u64,
    pub l: u64,
    pub c: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Artifact {
    pub id: String,          // art-<ulid>
    pub ts: String,          // RFC3339

    pub domain: String,      // domain.shared.world
    pub store: String,       // store token id

    pub kind: String,        // receipt | savepoint | agreement
    pub action: String,      // world.block.set.v1
    pub coord6: Coord6,
    pub reason: String,      // allow:* or deny:*

    #[serde(default)]
    pub refs: serde_json::Value,
}
