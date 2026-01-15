// src/api/chat.rs
use std::sync::Arc;
use warp::Filter;
use uuid::Uuid;

use crate::context::RuntimeContext;
use crate::id::DomainId;

use serde::Deserialize;

#[derive(Deserialize)]
struct ChatPost {
    body: String,
}

pub fn routes(
    runtime: Arc<RuntimeContext>,
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {

    let observe = warp::path!("observe" / "domain" / String / "chat")
        .and(warp::get())
        .and(with_runtime(runtime.clone()))
        .and_then(handle_observe_chat);

    let post = warp::path!("command" / "domain" / String / "chat")
        .and(warp::post())
        .and(warp::body::json())
        .and(with_runtime(runtime))
        .and_then(handle_post_chat);

    observe.or(post)
}

fn with_runtime(
    runtime: Arc<RuntimeContext>,
) -> impl Filter<Extract = (Arc<RuntimeContext>,), Error = std::convert::Infallible> + Clone {
    warp::any().map(move || runtime.clone())
}

async fn handle_post_chat(
    domain_id_raw: String,
    chat_post: ChatPost,
    runtime: Arc<RuntimeContext>,
) -> Result<impl warp::Reply, warp::Rejection> {

    let uuid = Uuid::parse_str(&domain_id_raw)
        .map_err(|_| warp::reject::not_found())?;

    let domain_id = DomainId(uuid);

    let message = runtime
        .chat()
        .append(domain_id, chat_post.body);

    Ok(warp::reply::json(&message))
}

async fn handle_observe_chat(
    domain_id_raw: String,
    runtime: Arc<RuntimeContext>,
) -> Result<impl warp::Reply, warp::Rejection> {

    let uuid = Uuid::parse_str(&domain_id_raw)
        .map_err(|_| warp::reject::not_found())?;

    let domain_id = DomainId(uuid);

    let room = runtime.chat().get_or_create(domain_id);

    Ok(warp::reply::json(&room))
}
