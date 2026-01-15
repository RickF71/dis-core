// src/spine/cube.rs

use serde::Serialize;

#[derive(Debug, Clone, Serialize)]
#[serde(bound = "
    T1: Serialize,
    T2: Serialize,
    T3: Serialize,
    T4: Serialize,
    T5: Serialize,
    T6: Serialize
")]
pub struct SpineCube<T1, T2, T3, T4, T5, T6> {
    pub nullus: T1,
    pub aether: T2,
    pub terra:  T3,
    pub numen:  T4,
    pub lima:   T5,
    pub corporeal: T6,
}
