// src/spine/projection.rs

/// A lawful projection from A → B.
/// The absence of an impl means the projection is forbidden.
pub trait Project<A, B> {
    fn project(a: A) -> B;
}


// src/spine/projection.rs (continued)

use crate::spine::numen::Numen;
use crate::spine::lima::Lima;

/// φ projection: Numen → Lima
impl Project<Numen, Lima> for () {
    fn project(n: Numen) -> Lima {
        Lima {
            echo: n.salience,
        }
    }
}
