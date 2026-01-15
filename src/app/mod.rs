use std::sync::Arc;
use warp::Filter;

use crate::context::RuntimeContext;
use crate::store::Store;

/// Node entrypoint.
/// Wires node-facing adapters only.
/// Defines no authority, policy, or meaning.
pub async fn run(port: u16) {
    println!("dis-core node listening on http://127.0.0.1:{port}");

    let store = Store::open().await.expect("Store::open");

    // Process-local coordination state (shared)
    let runtime = Arc::new(RuntimeContext::new(store.clone()));


    let http = crate::api::node::routes()
        .or(crate::api::domain::routes(runtime.clone()))
        .or(crate::api::chat::routes(runtime.clone()));

    let ws = crate::ws::routes::ws_routes(runtime.clone());

    let cors = warp::cors()
        .allow_any_origin()
        .allow_methods(vec!["GET", "POST", "OPTIONS"])
        .allow_headers(vec!["content-type"]);

    let routes = http
        .or(ws)
        .with(cors)
        .with(warp::log("dis.node"));

    warp::serve(routes)
        .run(([127, 0, 0, 1], port))
        .await;
}
