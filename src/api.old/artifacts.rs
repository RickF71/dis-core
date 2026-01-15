// src/api/artifacts.rs
use warp::Filter;
use crate::store::Store;

#[derive(Debug, serde::Deserialize)]
struct TailQuery {
    #[serde(default)]
    domain: Option<String>,
    #[serde(default)]
    after: Option<String>,
    #[serde(default = "default_limit")]
    limit: usize,
}

fn default_limit() -> usize { 200 }

pub fn routes(store: Store) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    let store_filter = warp::any().map(move || store.clone());

    let tail = warp::path!("api" / "artifacts" / "tail")
        .and(warp::get())
        .and(warp::query::<TailQuery>())
        .and(store_filter.clone())
        .and_then(handle_tail);

    let get = warp::path!("api" / "artifacts" / "get" / String)
        .and(warp::get())
        .and(store_filter.clone())
        .and_then(handle_get);

    tail.or(get)
}

async fn handle_tail(q: TailQuery, store: Store) -> Result<impl warp::Reply, warp::Rejection> {
    let limit = q.limit.clamp(1, 2000);
    let items = store
        .tail(q.domain.as_deref(), q.after.as_deref(), limit)
        .await
        .map_err(|_| warp::reject::not_found())?;

    Ok(warp::reply::json(&serde_json::json!({
        "items": items,
        "next_after": items.last().map(|a| a.id.clone()),
        "limit": limit
    })))
}

async fn handle_get(id: String, store: Store) -> Result<impl warp::Reply, warp::Rejection> {
    let item = store.get(&id).await.map_err(|_| warp::reject::not_found())?;
    match item {
        Some(a) => Ok(warp::reply::json(&a)),
        None => Err(warp::reject::not_found()),
    }
}
