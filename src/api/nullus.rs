// src/api/nullus.rs
use warp::Filter;

pub fn routes() -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path!("api" / "nullus")
        .map(|| "nullus ok")
}
