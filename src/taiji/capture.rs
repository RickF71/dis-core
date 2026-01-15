use crate::context::RuntimeContext;
use super::snapshot::*;

pub fn capture_snapshot(runtime: &RuntimeContext) -> RuntimeSnapshot {
    let observation = runtime.observation();
    let obs = observation.read().unwrap();
    let frame = obs.current_frame();


    RuntimeSnapshot {
        sequence: frame.sequence,
    }
}
