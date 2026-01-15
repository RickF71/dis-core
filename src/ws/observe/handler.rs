// src/ws/observe/handler.rs

use std::sync::Arc;
use std::time::Duration;

use futures_util::SinkExt; // required for `.send()`
use tokio::time::sleep;
use warp::ws::{Message, WebSocket};

use crate::context::RuntimeContext;
use crate::ws::observe::messages::*;
use crate::ws::observe::snapshot::build_snapshot;

pub async fn handle_observe_ws(
    mut socket: WebSocket,
    runtime: Arc<RuntimeContext>,
) {
    println!("🟢 /ws/observe client connected");

    let mut tick: u64 = 0;

    loop {
        tick += 1;

        // --------------------------------------------
        // Build snapshot from runtime
        // --------------------------------------------

        let snapshot = build_snapshot(&runtime);
        let msg = ObserveMessage::Snapshot(snapshot);

        let json = match serde_json::to_string(&msg) {
            Ok(j) => j,
            Err(e) => {
                eprintln!("🔴 observe serialize error: {e}");
                break;
            }
        };

        // --------------------------------------------
        // Send snapshot
        // --------------------------------------------

        match socket.send(Message::text(json)).await {
            Ok(_) => {
                println!("📤 observe snapshot sent (tick={tick})");
            }
            Err(e) => {
                println!("🔴 observe send failed (client disconnected?): {e}");
                break;
            }
        }

        // --------------------------------------------
        // Poll interval
        // --------------------------------------------

        sleep(Duration::from_millis(250)).await;
    }

    println!("🟠 /ws/observe client disconnected");
}
