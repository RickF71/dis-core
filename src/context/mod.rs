// src/context/mod.rs
use std::sync::{Arc, RwLock, Mutex};

use crate::store::Store;
use crate::context::observation::ObservationState;
use crate::runtime::totem::Totem;
use crate::runtime::totem_presence::TotemPresence;
use crate::chat::ChatRegistry;


pub mod observation;

#[derive(Clone)]
pub struct RuntimeContext {
    observation: Arc<RwLock<ObservationState>>,
    store: Store,

    // Process-local totem presence (no authority)
    totem_presence: Arc<Mutex<TotemPresence>>,

    // Authoritative totem (semantic commits)
    totem: Arc<Mutex<Totem>>,

    // Process-local chat state (no authority, no persistence)
    chat: Arc<ChatRegistry>,
}


impl RuntimeContext {
    pub fn new(store: Store) -> Self {
        Self {
            observation: Arc::new(RwLock::new(ObservationState::new())),
            store,
            totem_presence: Arc::new(Mutex::new(TotemPresence::new())),
            totem: Arc::new(Mutex::new(Totem::new())),
            chat: Arc::new(ChatRegistry::new()),
        }
    }

    pub fn observation(&self) -> Arc<RwLock<ObservationState>> {
        self.observation.clone()
    }

    pub fn store(&self) -> Store {
        self.store.clone()
    }

    pub fn totem(&self) -> Arc<Mutex<Totem>> {
        self.totem.clone()
    }

    pub fn totem_presence(&self) -> Arc<Mutex<TotemPresence>> {
        self.totem_presence.clone()
    }

    pub fn chat(&self) -> Arc<ChatRegistry> {
        self.chat.clone()
    }

}
