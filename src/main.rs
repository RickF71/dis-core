use std::sync::Arc;
use warp::Filter;


use triad::api;
use triad::context::RuntimeContext;
use triad::store::Store;
use triad::ws;

#[tokio::main]
async fn main() {
    eprintln!("[BOOT] initializing");

    // -------------------------------------------------
    // Persistent store
    // -------------------------------------------------
    eprintln!("[BOOT] opening store");
    let store = Store::open().await.expect("Store::open failed");
    eprintln!("[BOOT] store ready");

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
    let routes = api_routes.or(ws_routes).or(warp::path("ping").map(|| "pong"));

    let addr = ([127, 0, 0, 1], 8710);

    eprintln!("[BOOT] binding http listener {}:{}", addr.0[0], addr.1);

    // Bind now, log AFTER it succeeds, and shutdown cleanly on Ctrl-C.
    let (bound_addr, server) = warp::serve(routes).bind_with_graceful_shutdown(addr, async {
        // Wait for Ctrl-C
        let _ = tokio::signal::ctrl_c().await;
        eprintln!("[BOOT] shutdown requested (Ctrl-C)");
    });

    eprintln!("[BOOT] listener ready on http://{}", bound_addr);
    eprintln!("[BOOT] idle");

    server.await;

    eprintln!("[BOOT] shutdown complete");
}
