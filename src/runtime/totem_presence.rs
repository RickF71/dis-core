// src/runtime/totem_runtime.rs
use std::time::{Duration, Instant};

#[derive(Debug)]
pub struct TotemPresence {
    last_seen: Option<Instant>,
    since: Option<Instant>,
}


impl TotemPresence {
    pub fn new() -> Self {
        Self {
            last_seen: None,
            since: None,
        }
    }

    pub fn heartbeat(&mut self) {
        let now = Instant::now();
        if self.last_seen.is_none() {
            self.since = Some(now);
        }
        self.last_seen = Some(now);
    }

    pub fn is_present(&self, timeout: Duration) -> bool {
        self.last_seen
            .map(|t| t.elapsed() <= timeout)
            .unwrap_or(false)
    }

    pub fn last_seen_ms(&self) -> Option<u64> {
        self.last_seen.map(|t| t.elapsed().as_millis() as u64)
    }

    pub fn clear(&mut self) {
        self.last_seen = None;
        self.since = None;
    }
}
