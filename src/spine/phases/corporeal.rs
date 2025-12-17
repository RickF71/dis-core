use crate::domain::seat::Seat;
use crate::domain::domain_id::DomainId;
use crate::spine::clock::DisTick;

/// Final execution phase of the Layer6 pipeline.
/// This is the terminal corporeal projection step.
pub fn execute_final_layer6_phase(
    seat: &mut Seat,
    dis_tick: DisTick,
) {
    seat.upsert_corporeal_projection(
        DomainId::self_domain(),
        dis_tick,
    );
}
