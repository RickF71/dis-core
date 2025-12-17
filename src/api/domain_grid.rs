// src/api/domain_grid.rs

use warp::Filter;
use crate::api::grid::generate_full_grid;

pub fn routes() -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path!("api" / "domain" / String / "grid")
        .and(warp::get())
        .map(|_domain: String| {
            let grid = generate_full_grid();
            warp::reply::json(&grid)
        })
}
