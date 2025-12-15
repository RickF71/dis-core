use crate::spine::clock::DisTick;
use crate::domain::domain_id::DomainId;
use crate::domain::seat::SeatId;

#[derive(Debug, Clone)]
pub struct CorporealSeatProjection {
    /// Domain this projection exists in
    pub domain_id: DomainId,

    /// Owning pseat (never leaves home domain)
    pub owner_seat: SeatId,

    /// Local time of this seat within the domain
    pub seat_domain_tick: u64,

    /// DIS tick when this projection was last active
    pub last_active_dis_tick: DisTick,

    /// Domain-granted permissions (opaque for now)
    pub permissions: DomainPermissions,
}


#[derive(Debug, Clone)]
pub struct DomainPermissions {
    pub can_act: bool,
}
