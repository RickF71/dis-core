// src/domain/receipts/mod.rs
use uuid::Uuid;
use sha2::{Sha256, Digest};

pub fn new_receipt_id() -> String {
    format!("rcpt-{}", Uuid::new_v4())
}

pub fn hash_payload(payload: &str) -> String {
    let mut hasher = Sha256::new();
    hasher.update(payload.as_bytes());
    format!("{:x}", hasher.finalize())
}
