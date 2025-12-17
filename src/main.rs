use warp::Filter;

use crate::kernel::Kernel;

mod bootstrap;
mod context;
mod api;
mod kernel;

#[tokio::main]
async fn main() {
    println!("dis_core runtime starting on http://localhost:8787");

    // -------------------------
    // Kernel = authority boundary
    // -------------------------
    let kernel = Kernel::new();

    // -------------------------
    // API routes
    // -------------------------
    let commit_routes = api::commit::routes(kernel.clone());

    // You can add more routes here later:
    // let status_routes = api::status::routes();
    // let domain_routes = api::domain::routes(kernel.clone());

    let routes = commit_routes;

    // -------------------------
    // Server
    // -------------------------
    warp::serve(routes)
        .run(([127, 0, 0, 1], 8787))
        .await;
}
