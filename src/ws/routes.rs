// src/ws/routes.rs

use std::sync::Arc;
use warp::Filter;

use crate::context::RuntimeContext;

use crate::ws::observe::handler::handle_observe_ws;
use crate::ws::command::handler::handle_command_ws;
use crate::ws::totem::handler::handle_totem_ws;

use futures_util::StreamExt;


/// Node-facing WebSocket routes.
/// Transport only. No authority.
pub fn ws_routes(
    runtime: Arc<RuntimeContext>,
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {

    // Inject RuntimeContext into all WS handlers
    let runtime_filter = warp::any().map(move || runtime.clone());

    // ---------------------------------------------
    // Observer WebSocket (read-only projection)
    // ws://host/ws/observe
    // ---------------------------------------------
    let observe = warp::path!("ws" / "observe")
        .and(warp::ws())
        .and(runtime_filter.clone())
        .map(|ws: warp::ws::Ws, runtime| {
            ws.on_upgrade(move |socket| {
                handle_observe_ws(socket, runtime)
            })
        });

    // ---------------------------------------------
    // Command WebSocket (intent submission)
    // ws://host/ws/command
    // ---------------------------------------------
    let command = warp::path!("ws" / "command")
        .and(warp::ws())
        .and(runtime_filter.clone())
        .map(|ws: warp::ws::Ws, runtime| {
            ws.on_upgrade(move |socket| {
                handle_command_ws(socket, runtime)
            })
        });

    // ---------------------------------------------
    // Totem WebSocket (presence & consent surface)
    // ws://host/ws/totem
    // ---------------------------------------------
    let totem = warp::path!("ws" / "totem")
        .and(warp::ws())
        .and(runtime_filter)
        .map(|ws: warp::ws::Ws, runtime: Arc<RuntimeContext>| {
            ws.on_upgrade(move |ws| async move {
                let (_tx, mut rx) = ws.split();

                while let Some(Ok(msg)) = rx.next().await {
                    handle_totem_ws(msg, runtime.clone()).await;
                }
            })
        });

    observe.or(command).or(totem)
}
