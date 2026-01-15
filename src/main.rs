use std::sync::Arc;
use warp::Filter;

use triad::context::RuntimeContext;
use triad::store::Store;
use triad::api;
use triad::ws;

#[tokio::main]
async fn main() {
    println!("BOOT: starting DIS server on 127.0.0.1:8710");

    // -------------------------------------------------
    // Persistent store
    // -------------------------------------------------
    let store = Store::open()
        .await
        .expect("Store::open failed");

    // -------------------------------------------------
    // Runtime context (process-local)
    // -------------------------------------------------
    let runtime = Arc::new(RuntimeContext::new(store));

    // -------------------------------------------------
    // HTTP API routes (chat, domain, observe, etc.)
    // -------------------------------------------------
    let api_routes = api::routes(runtime.clone());

    // -------------------------------------------------
    // WebSocket routes (/ws/observe, /ws/command, /ws/totem)
    // -------------------------------------------------
    let ws_routes = ws::ws_routes(runtime.clone());

    // -------------------------------------------------
    // Root routes
    // -------------------------------------------------
    let routes = api_routes
        .or(ws_routes)
        .or(warp::path("ping").map(|| "pong"));

    warp::serve(routes)
        .run(([127, 0, 0, 1], 8710))
        .await;
}
