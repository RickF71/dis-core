// src/api/node.rs
//
// Node-level observation APIs.
// Non-authoritative. No domain routing.

use warp::Filter;
use serde::Serialize;

#[derive(Serialize)]
struct NodeStatus {
    runtime: &'static str,
    version: &'static str,
}

pub fn routes(
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path!("api" / "node" / "status")
        .and(warp::get())
        .map(|| {
            warp::reply::json(&NodeStatus {
                runtime: "dis-core",
                version: env!("CARGO_PKG_VERSION"),
            })
        })
}
