// src/domain/http/domain_routes.rs
//
// Phase 5 adapter (Human / Network surface).
// This file exposes domain information over HTTP.
// It must not mutate domain state or encode authority.

use warp::Filter;
use serde::Serialize;

//
// -----------------------------
// Public Domain Surface (DTO)
// -----------------------------
//

#[derive(Debug, Serialize)]
pub struct DomainSurface {
    pub meta: DomainMeta,
    pub status: DomainStatus,
}

#[derive(Debug, Serialize)]
pub struct DomainMeta {
    pub id: String,
    pub version: String,
}

#[derive(Debug, Serialize)]
pub struct DomainStatus {
    pub recognized: bool,
    pub frozen: bool,
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
            let surface = build_stub_surface(&domain_id);
            warp::reply::json(&surface)
        })
}

//
// -----------------------------
// Stub Builder
// -----------------------------
//

fn build_stub_surface(domain_id: &str) -> DomainSurface {
    DomainSurface {
        meta: DomainMeta {
            id: domain_id.to_string(),
            version: env!("CARGO_PKG_VERSION").to_string(),
        },
        status: DomainStatus {
            recognized: false,
            frozen: false,
        },
    }
}
