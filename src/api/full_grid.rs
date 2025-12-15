use warp::Filter;
use crate::api::grid::generate_full_grid;

pub fn routes() -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path!("api" / "grid")
        .and(warp::get())
        .map(|| {
            let grid = generate_full_grid();
            warp::reply::json(&grid)
        })
}
