use warp::Filter;
use warp::cors;

mod api;
mod bootstrap;
mod context;

#[tokio::main]
async fn main() {
    println!("dis_spine runtime starting on http://localhost:8787");

    let cors = cors()
        .allow_any_origin()
        .allow_methods(vec!["GET"])
        .allow_headers(vec!["Content-Type"]);

    let routes = api::status::routes()
        .or(api::domain_context::routes())
        .or(api::domain_grid::routes())
        .with(cors);

    warp::serve(routes)
        .run(([127, 0, 0, 1], 8787))
        .await;
}
