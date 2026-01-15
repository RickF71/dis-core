// src/chat/mod.rs

use std::collections::HashMap;
use std::sync::RwLock;

use serde::Serialize;

use crate::id::DomainId;

// ============================================================
// Chat Message — Semantic Event
// ============================================================

#[derive(Clone, Debug, Serialize)]
pub struct ChatMessage {
    /// Monotonic identifier within a room (projection-scoped for now)
    pub id: u64,

    /// Message body (no formatting assumptions)
    pub body: String,
}

// ============================================================
// Chat Room — Materialized Projection
// ============================================================

#[derive(Clone, Debug, Serialize)]
pub struct ChatRoom {
    /// Ordered messages for this domain
    pub messages: Vec<ChatMessage>,

    /// Logical tick / version for projection invalidation
    pub tick: u64,
}

impl ChatRoom {
    pub fn new() -> Self {
        Self {
            messages: Vec::new(),
            tick: 0,
        }
    }
}

// ============================================================
// Chat Registry — Domain-Indexed Store
// ============================================================

pub struct ChatRegistry {
    rooms: RwLock<HashMap<DomainId, ChatRoom>>,
}

impl ChatRegistry {
    pub fn new() -> Self {
        Self {
            rooms: RwLock::new(HashMap::new()),
        }
    }

    /// Get a cloned snapshot of a room (for HTTP / observe projection)
    pub fn get_or_create(&self, domain: DomainId) -> ChatRoom {
        let mut rooms = self.rooms.write().unwrap();
        rooms
            .entry(domain)
            .or_insert_with(ChatRoom::new)
            .clone()
    }

    /// Append a new chat message to a domain room.
    /// Returns the semantic ChatMessage event.
    pub fn append(&self, domain: DomainId, body: String) -> ChatMessage {
        let mut rooms = self.rooms.write().unwrap();
        let room = rooms.entry(domain).or_insert_with(ChatRoom::new);

        let message = ChatMessage {
            id: room.tick,
            body,
        };

        room.messages.push(message.clone());
        room.tick += 1;

        message
    }
}
