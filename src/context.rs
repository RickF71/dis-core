// src/context.rs

use std::sync::{Arc, RwLock};

use dis_core::spine::clock::DisClock;
use serde::Serialize;

use crate::bootstrap::ColorDef;

#[derive(Clone)]
pub struct RuntimeContext {
    pub clock: Arc<RwLock<DisClock>>,
}

impl RuntimeContext {
    pub fn new() -> Self {
        Self {
            clock: Arc::new(RwLock::new(DisClock::new())),
        }
    }
}

#[derive(Serialize)]
pub struct DomainContext {
    pub domain: DomainInfo,
    pub bootstrap: Bootstrap,
    pub projection: Option<Projection>,
}

/// Descriptive only — rendered for UI / API consumers
#[derive(Serialize)]
pub struct DomainInfo {
    pub id: String,
    pub focus: String,
    pub parent: Option<String>,
}

#[derive(Serialize)]
pub struct Bootstrap {
    pub colors: std::collections::HashMap<String, ColorDef>,
    pub layers: std::collections::HashMap<String, String>,
}

#[derive(Serialize)]
pub struct Projection {
    pub r#type: String,
    pub axes: Vec<String>,
}

// build_domain_context and DomainAuthority usage commented out for clean compile
