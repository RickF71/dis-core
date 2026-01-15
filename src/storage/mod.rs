// src/storage/mod.rs
use anyhow::Result;

pub trait ReceiptLedger: Send + Sync {
    fn append(&mut self, record: &[u8]) -> Result<()>;
    fn iter(&self) -> Result<Box<dyn Iterator<Item = Vec<u8>> + Send>>;
}

pub trait ArtifactStore: Send + Sync {
    fn put(&mut self, bytes: &[u8]) -> Result<String>;      // returns content hash (string for now)
    fn get(&self, hash: &str) -> Result<Option<Vec<u8>>>;
}

pub struct TotemStorage {
    pub ledger: Box<dyn ReceiptLedger>,
    pub artifacts: Box<dyn ArtifactStore>,
}
