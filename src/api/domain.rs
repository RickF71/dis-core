// src/api/domain.rs

use warp::Filter;
use serde::{Serialize, Deserialize};
use std::sync::Arc;
use std::time::Duration;

use crate::context::RuntimeContext;
use crate::runtime::commit::CommitKind;

#[derive(Deserialize)]
struct ApplyRequest {
    intent: String,
}

#[derive(Serialize)]
struct DomainStatus {
    domain: String,
    totem_present: bool,
}

pub fn routes(
    ctx: Arc<RuntimeContext>,
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {

    let ctx_filter = warp::any().map(move || ctx.clone());

    // ----------------------------
    // GET /api/domain/{id}/status
    // ----------------------------
    let status = warp::path!("api" / "domain" / String / "status")
        .and(warp::get())
        .and(ctx_filter.clone())
        .map(|domain_id: String, ctx: Arc<RuntimeContext>| {
            let present = {
                let presence_handle = ctx.totem_presence();
                let presence = presence_handle
                    .lock()
                    .expect("totem_presence poisoned");

                presence.is_present(Duration::from_secs(5))
            };

            warp::reply::json(&DomainStatus {
                domain: domain_id,
                totem_present: present,
            })
        });

    // ----------------------------
    // POST /api/domain/{id}/apply
    // ----------------------------
    let apply = warp::path!("api" / "domain" / String / "apply")
        .and(warp::post())
        .and(warp::body::json())
        .and(ctx_filter.clone())
        .map(|_domain_id: String, req: ApplyRequest, ctx: Arc<RuntimeContext>| {
            let receipt = {
                let totem_handle = ctx.totem();
                let mut totem = totem_handle
                    .lock()
                    .expect("totem poisoned");

                totem.commit(CommitKind::MemoryAppend, req.intent)
            };

            warp::reply::json(&receipt)
        });

    status.or(apply)
}
