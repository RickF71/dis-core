// src/ws/command/handler.rs

use std::sync::Arc;

use warp::ws::{Message, WebSocket};
use futures_util::{StreamExt, SinkExt};

use crate::context::RuntimeContext;
use crate::ws::command::messages::{
    CommandMessage,
    CommandResponse,
};

pub async fn handle_command_ws(
    socket: WebSocket,
    _runtime: Arc<RuntimeContext>,
) {
    let (mut tx, mut rx) = socket.split();

    while let Some(Ok(msg)) = rx.next().await {
        if !msg.is_text() {
            continue;
        }

        let parsed = serde_json::from_str::<CommandMessage>(
            msg.to_str().unwrap(),
        );

        let response = match parsed {
            Ok(_command) => {
                // no authority, no mutation yet
                CommandResponse::Accepted {
                    intent_id: "intent-temp-id".to_string(),
                }
            }
            Err(err) => {
                CommandResponse::Denied {
                    intent_id: "unknown".to_string(),
                    reason: format!("invalid command: {}", err),
                }
            }
        };

        let _ = tx
            .send(Message::text(
                serde_json::to_string(&response).unwrap(),
            ))
            .await;
    }
}
