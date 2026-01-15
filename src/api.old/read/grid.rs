use warp::Filter;
use crate::api::read::grid_model::generate_full_grid;

pub fn routes() -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    let root = warp::path!("api" / "grid")
        .and(warp::get())
        .map(|| warp::reply::json(&generate_full_grid()));

    let by_domain = warp::path!("api" / "domain" / String / "grid")
        .and(warp::get())
        .map(|_domain: String| warp::reply::json(&generate_full_grid()));

    root.or(by_domain)
}