use crate::id::{DomainId, SeatId};
use super::{FacetId};

/// Nullus primary interface
pub trait TotemIfc {
    fn facet_id(&self) -> FacetId;
    fn domain_id(&self) -> DomainId;

    // "Existence / binding" primitives (stubs for now)
    fn totem_hello(&self) -> TotemHello;
    fn bind_actor(&self, _actor: ActorRef) -> BindResult;
}

/// Aether primary interface
pub trait SeatIfc {
    fn facet_id(&self) -> FacetId;
    fn domain_id(&self) -> DomainId;

    fn mint_seat(&mut self) -> SeatId;
    fn seat_present(&self, seat: SeatId) -> bool;
    fn enter_seat(&mut self, seat: SeatId);
    fn leave_seat(&mut self, seat: SeatId);
}

/// Terra primary interface
pub trait DomainIfc {
    fn facet_id(&self) -> FacetId;
    fn domain_id(&self) -> DomainId;

    fn domain_surface(&self) -> DomainSurface;
}

/// Numen primary interface
pub trait ResolutionIfc {
    fn facet_id(&self) -> FacetId;
    fn domain_id(&self) -> DomainId;

    fn resolve(&mut self, _req: ResolveRequest) -> ResolveResult;
}

/// Lima primary interface
pub trait MeaningIfc {
    fn facet_id(&self) -> FacetId;
    fn domain_id(&self) -> DomainId;

    fn annotate(&mut self, _note: Note) -> ();
    fn meaning_surface(&self) -> MeaningSurface;
}

/// Corporeal primary interface
pub trait InteractionIfc {
    fn facet_id(&self) -> FacetId;
    fn domain_id(&self) -> DomainId;

    fn open_session(&mut self) -> SessionId;
    fn accept_input(&mut self, _sid: SessionId, _input: InputEvent) -> ();
    fn render(&self, _sid: SessionId) -> RenderFrame;

    /// Corporeal is the *only* facet allowed to mint child domains.
    fn mint_child_domain(&self) -> DomainId;
}

// -----------------------------
// Minimal types (stubs)
// -----------------------------

#[derive(Debug, Clone)]
pub struct ActorRef {
    pub opaque: String,
}

#[derive(Debug, Clone)]
pub struct TotemHello {
    pub domain_id: DomainId,
    pub facet_id: FacetId,
    pub state: &'static str,
}

#[derive(Debug, Clone)]
pub struct BindResult {
    pub ok: bool,
}

#[derive(Debug, Clone)]
pub struct DomainSurface {
    pub id: DomainId,
    pub facet_id: FacetId,
    pub reachable: bool,
}

#[derive(Debug, Clone)]
pub struct ResolveRequest {
    pub claim: String,
}

#[derive(Debug, Clone)]
pub struct ResolveResult {
    pub decided: bool,
}

#[derive(Debug, Clone)]
pub struct Note {
    pub text: String,
}

#[derive(Debug, Clone)]
pub struct MeaningSurface {
    pub facet_id: FacetId,
    pub notes_count: usize,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct SessionId(pub u64);

#[derive(Debug, Clone)]
pub struct InputEvent {
    pub kind: String,
    pub data: String,
}

#[derive(Debug, Clone)]
pub struct RenderFrame {
    pub ok: bool,
    pub message: String,
}
