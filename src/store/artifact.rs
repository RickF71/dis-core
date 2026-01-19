use serde::{Deserialize, Serialize};
use crate::runtime::coord6::Coord6;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Artifact {
    pub id: String,          // art-<ulid>

    // Actor-asserted observation time.
    // DIS-core does not interpret or trust this value.

    pub observed_at: Option<String>, // actor-asserted RFC3339

    pub domain: String,      // domain.shared.world
    pub store: String,       // store token id

    pub kind: String,        // receipt | savepoint | agreement
    pub action: String,      // world.block.set.v1
    pub coord6: Coord6,
    pub reason: String,      // allow:* or deny:*

    #[serde(default)]
    pub refs: serde_json::Value,
}
