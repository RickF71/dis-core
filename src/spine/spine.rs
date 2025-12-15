use crate::spine::{
    Layer6,
    capsule::Capsule,
    node::Node,
    echo_store::Stored,
};

#[derive(Debug, Clone)]
pub enum SpineYield {
    Exit(Capsule<Node>),
    Stored(Stored),
}


/// Run the full Layer6 spine.
/// This function is the ONLY legal path from Nullus to Corporeal.
pub fn run(mut c: Capsule<Node>) -> SpineYield {
    // ---- Aether (connectivity / returnability)
    c = c.tick_to(Layer6::Aether);

    // ---- Terra (presence / locality)
    c = c.tick_to(Layer6::Terra);

    // ---- Numen (noticing / salience)
    c = c.tick_to(Layer6::Numen);

    // For now: trivial salience
    let salience: u8 = 1;

    // ---- Lima (processing / incorporation)
    c = c.tick_to(Layer6::Lima);

    // ---- Decision boundary (co-φ)
    if salience == 0 {
        // Memory captured below Corporeal
        SpineYield::Stored(Stored {
            echo: 0,
        })
    } else {
        // ---- Corporeal (exit)
        c = c.tick_to(Layer6::Corporeal);
        SpineYield::Exit(c)
    }
}
