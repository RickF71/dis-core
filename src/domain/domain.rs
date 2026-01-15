// src/domain/domain.rs
use warp::Filter;
use serde::Serialize;

use crate::domain::lattice::DomainLattice;
use crate::id::DomainId;

//
// -----------------------------
// Domain Ontology (CORE)
// -----------------------------
//
#[derive(Debug, Clone)]
pub struct Domain {
    pub id: DomainId,
    pub parent_id: Option<DomainId>,

    /// Declared lattice / structure (may be partial or empty)
    pub lattice: Option<DomainLattice>,
}

//
// -----------------------------
// Domain Surface (public face)
// -----------------------------
//

#[derive(Debug, Serialize)]
pub struct DomainSurface {
    pub meta: DomainMeta,
    pub lattice: Option<DomainLattice>,
    pub status: DomainStatus,
    pub interfaces: DomainInterfaces,
}

#[derive(Debug, Serialize)]
pub struct DomainMeta {
    pub id: String,
    pub parent_id: Option<String>,
    pub version: String,
}

#[derive(Debug, Serialize)]
pub struct DomainStatus {
    pub recognized: bool,
    pub frozen: bool,
}

#[derive(Debug, Serialize)]
pub struct DomainInterfaces {
    pub interfaces: Vec<String>,
    pub endpoints: std::collections::HashMap<String, String>,
    pub permissions: std::collections::HashMap<String, bool>,
}

//
// -----------------------------
// Routes
// -----------------------------
//

pub fn routes(
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path!("api" / "domain" / String)
        .and(warp::get())
        .map(|domain_id: String| {
            let surface = build_domain_surface(&domain_id);
            warp::reply::json(&surface)
        })

}

//
// -----------------------------
// Builder
// -----------------------------
//

fn build_domain_surface(domain_id: &str) -> DomainSurface {
    let meta = DomainMeta {
        id: domain_id.to_string(),
        parent_id: None,
        version: env!("CARGO_PKG_VERSION").to_string(),
    };

    let status = DomainStatus {
        recognized: false,
        frozen: false,
    };

    let mut endpoints = std::collections::HashMap::new();
    endpoints.insert(
        "self".to_string(),
        format!("/api/domain/{}", domain_id),
    );

    let mut permissions = std::collections::HashMap::new();
    permissions.insert("can_commit".to_string(), !status.frozen);

    let interfaces = DomainInterfaces {
        interfaces: vec!["http".to_string()],
        endpoints,
        permissions,
    };

    DomainSurface {
        meta,
        lattice: None,
        status,
        interfaces,
    }
}
