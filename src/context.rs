use std::sync::{Arc, RwLock};
use dis_spine::spine::clock::DisClock;

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
use serde::Serialize;

#[derive(Serialize)]
pub struct DomainContext {
    pub domain: DomainInfo,
    pub bootstrap: Bootstrap,
    pub projection: Option<Projection>,
}

#[derive(Serialize)]
pub struct DomainInfo {
    pub id: String,
    pub layer: String,
    pub parent: Option<String>,
}

#[derive(Serialize)]
pub struct Bootstrap {
    pub colors: std::collections::HashMap<String, String>,
    pub layers: std::collections::HashMap<String, String>,
}

#[derive(Serialize)]
pub struct Projection {
    pub r#type: String,
    pub axes: Vec<String>,
}

use crate::bootstrap;

pub fn build_domain_context(domain_id: &str) -> DomainContext {
    let colors = bootstrap::load_colors().colors;
    let layers = bootstrap::load_layers().layers;

    DomainContext {
        domain: DomainInfo {
            id: domain_id.to_string(),
            layer: "terra".to_string(), // for now
            parent: Some("domain.aether".to_string()),
        },
        bootstrap: Bootstrap {
            colors,
            layers,
        },
        projection: Some(Projection {
            r#type: "cube".to_string(),
            axes: vec![
                "nullus", "aether", "terra",
                "numen", "lima", "corporeal"
            ].into_iter().map(String::from).collect(),
        }),
    }
}
