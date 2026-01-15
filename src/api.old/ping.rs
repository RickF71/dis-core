// src/api/ping.rs
use warp::Filter;

pub fn routes() -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path!("api" / "ping")
        .and(warp::get())
        .map(|| "pong")
}
