use crate::id::DomainId;
use super::{FacetMeta, FacetKind, FacetId};
use super::interfaces::*;

#[derive(Debug)]
pub struct NullusFacet {
    pub meta: FacetMeta,
}

impl NullusFacet {
    /// Root facet in a domain chain.
    pub fn mint_root(domain_id: DomainId) -> Self {
        Self {
            meta: FacetMeta::new(domain_id, FacetKind::Nullus, None),
        }
    }

    /// Mint Aether from Nullus (explicit, linear).
    pub fn mint_aether(&self) -> super::aether::AetherFacet {
        super::aether::AetherFacet::mint_from_parent(self)
    }
}

impl TotemIfc for NullusFacet {
    fn facet_id(&self) -> FacetId { self.meta.id }
    fn domain_id(&self) -> DomainId { self.meta.domain_id }

    fn totem_hello(&self) -> TotemHello {
        TotemHello {
            domain_id: self.meta.domain_id,
            facet_id: self.meta.id,
            state: "unbound",
        }
    }

    fn bind_actor(&self, _actor: ActorRef) -> BindResult {
        // no-op stub: existence anchoring only
        BindResult { ok: true }
    }
}
