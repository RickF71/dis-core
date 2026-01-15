// src/api/mod.rs
//
// Canonical API surface for dis-core.
// APIs are routing layers only.
// No authority, no persistence, no meaning.

pub mod node;
pub mod domain;
pub mod chat;

use std::sync::Arc;
use warp::Filter;

use crate::context::RuntimeContext;


pub fn routes(
    runtime: Arc<RuntimeContext>,
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    chat::routes(runtime)
}
