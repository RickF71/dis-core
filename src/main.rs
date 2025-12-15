use crate::context::RuntimeContext;
use dis_spine::spine::clock::SpinePhase;
use warp::Filter;
use warp::cors;
use warp::Reply;
use serde_json::json;

mod bootstrap;
mod context;
mod api;

fn load_root_html() -> &'static str {
    include_str!("bootstrap/root.html")
}

#[tokio::main]
async fn main() {
    println!("dis_spine runtime starting on http://localhost:8787");

    let cors = cors()
        .allow_any_origin()
        .allow_methods(vec!["GET"])
        .allow_headers(vec!["Content-Type"]);

    // Initialize shared runtime context
    let context = RuntimeContext::new();
    let clock_filter = warp::any().map(move || context.clone());

    // Handshake payload for root endpoint
    let handshake = json!({
        "name": "dis_spine",
        "role": "spine",
        "status": "running",
        "api": "/api",
        "version": "0.1.0"
    });

    // Root '/' endpoint with content negotiation
    let root = warp::path::end()
        .and(warp::get())
        .and(warp::header::optional::<String>("accept"))
        .map(move |accept: Option<String>| {
            let wants_html = accept
                .as_deref()
                .map(|a| a.contains("text/html"))
                .unwrap_or(false);

            if wants_html {
                warp::reply::html(load_root_html())
                    .into_response()
            } else {
                warp::reply::json(&handshake)
                    .into_response()
            }
        });

    // /api/status endpoint
    let api_status = warp::path!("api" / "status")
        .and(warp::get())
        .and(clock_filter.clone())
        .map(|ctx: RuntimeContext| {
            let clock = ctx.clock.read().unwrap();
            warp::reply::json(&serde_json::json!({
                "service": "dis_spine",
                "role": "spine",
                "version": "0.1.0",
                "status": "running",
                "dis_tick": clock.dis_tick,
                "phase": format!("{:?}", clock.phase),
                "commit_allowed": clock.commit_allowed()
            }))
        });

    // /api/domain/:id/context
    let domain_context = warp::path!("api" / "domain" / String / "context")
        .and(warp::get())
        .map(|domain_id: String| {
            let ctx = context::build_domain_context(&domain_id);
            warp::reply::json(&ctx)
        });

    let routes = root
        .or(api_status)
        .or(api::full_grid::routes())   // ✅ THIS FIXES /api/grid
        .or(api::domain_grid::routes())
        .or(domain_context)
        .or(api::nullus::routes())
        .with(cors);

    warp::serve(routes)
        .run(([127, 0, 0, 1], 8787))
        .await;
}
