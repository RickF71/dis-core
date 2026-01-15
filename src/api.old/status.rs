// src/api/status.rs

use warp::Filter;
use serde::Serialize;

use std::sync::Arc;
use crate::context::RuntimeContext;

#[derive(Serialize)]
struct StatusResponse {
    runtime: &'static str,
    version: &'static str,

    /// Monotonic observation counter.
    /// This is not wall-clock time and carries no authority.
    observation_sequence: u64,

    /// UI anchoring only — carries no authority.
    #[serde(skip_serializing_if = "Option::is_none")]
    domains: Option<Vec<DomainStub>>,
}

#[derive(Serialize)]
struct DomainStub {
    id: String,
    layer: &'static str,
}

pub fn routes(
    ctx: Arc<RuntimeContext>,
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    let ctx_filter = warp::any().map(move || ctx.clone());

    warp::path!("api" / "status")
        .and(warp::get())
        .and(ctx_filter)
        .map(|ctx: Arc<RuntimeContext>| {
            let frame = ctx
                .observation()
                .read()
                .expect("observation poisoned")
                .current_frame();

            warp::reply::json(&StatusResponse {
                runtime: "dis-core",
                version: env!("CARGO_PKG_VERSION"),
                observation_sequence: frame.sequence,

                // Minimal domain anchor for Finagler
                domains: Some(vec![
                    DomainStub {
                        id: "domain.local".into(),
                        layer: "corporeal",
                    }
                ]),
            })
        })
}
