// src/api/domain_grid.rs

use warp::Filter;
use dis_spine::spine::Layer6;
use crate::api::grid::generate_full_grid;

fn parse_layer(s: &str) -> Option<Layer6> {
    match s {
        "nullus" => Some(Layer6::Nullus),
        "aether" => Some(Layer6::Aether),
        "terra" => Some(Layer6::Terra),
        "numen" => Some(Layer6::Numen),
        "lima" => Some(Layer6::Lima),
        "corporeal" => Some(Layer6::Corporeal),
        _ => None,
    }
}

pub fn routes() -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path!("api" / "domain" / String / "grid")
        .and(warp::get())
        .map(|domain: String| {
            let layer = parse_layer(&domain)
                .expect("invalid domain");

            let grid = generate_full_grid();
            warp::reply::json(&grid)
        })
}
