// src/taiji/projection.rs

use crate::context::RuntimeContext;
use crate::context::observation::ObservationFrame;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ObservationProjection {
    pub frame: ObservationFrame,
    pub domain: String,
    pub tokens: Vec<ObservedToken>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ObservedToken {
    pub id: String,
    pub domain: String,
}

impl ObservationProjection {
    /// Capture only the observation frame (sync, no await)
    fn capture_frame(runtime: &RuntimeContext) -> ObservationFrame {
        let observation = runtime.observation();
        let obs = observation.read().unwrap();
        obs.current_frame()
    }

    /// Full capture (async, Send-safe)
    pub async fn capture(
        runtime: &RuntimeContext,
        domain: &str,
    ) -> Self {
        // 1) observation coordinate (SYNC)
        let frame = Self::capture_frame(runtime);

        // 2) artifacts == tokens (ASYNC)
        let tokens = runtime
            .store()
            .tail(Some(domain), None, 2000)
            .await
            .unwrap_or_default()
            .into_iter()
            .map(|a| ObservedToken {
                id: a.id.clone(),
                domain: a.domain.clone(),
            })
            .collect();

        Self {
            frame,
            domain: domain.to_string(),
            tokens,
        }
    }
}
