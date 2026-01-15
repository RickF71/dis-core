// src/totem/trait.rs
use anyhow::Result;

#[derive(Clone, Debug)]
pub struct ReceiptBytes(pub Vec<u8>);

#[derive(Clone, Debug)]
pub struct IntentBytes(pub Vec<u8>);

pub trait Totem: Send + Sync {
    fn domain_id(&self) -> &str;

    /// Rebuild in-memory state from durable history.
    fn replay(&mut self) -> Result<()>;

    /// Apply an intent; must persist output and return the new receipt bytes.
    fn apply_intent(&mut self, intent: IntentBytes) -> Result<ReceiptBytes>;
}
