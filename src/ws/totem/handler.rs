use std::sync::Arc;
use warp::ws::Message;

use crate::context::RuntimeContext;

pub async fn handle_totem_ws(
    msg: Message,
    ctx: Arc<RuntimeContext>,
) {
    if msg.is_text() {
        let presence_handle = ctx.totem_presence();
        let mut presence = presence_handle
            .lock()
            .expect("totem_presence poisoned");

        presence.heartbeat();
    }

    if msg.is_close() {
        let presence_handle = ctx.totem_presence();
        let mut presence = presence_handle
            .lock()
            .expect("totem_presence poisoned");

        presence.clear();
    }
}
