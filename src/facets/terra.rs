use crate::id::DomainId;
use super::{FacetMeta, FacetKind, FacetId};
use super::interfaces::*;
use super::aether::AetherFacet;

#[derive(Debug)]
pub struct TerraFacet {
    pub meta: FacetMeta,
    reachable: bool,
}

impl TerraFacet {
    pub fn mint_from_parent(parent: &AetherFacet) -> Self {
        Self {
            meta: FacetMeta::new(parent.meta.domain_id, FacetKind::Terra, Some(parent.meta.id)),
            reachable: true,
        }
    }

    pub fn mint_numen(&self) -> super::numen::NumenFacet {
        super::numen::NumenFacet::mint_from_parent(self)
    }
}

impl DomainIfc for TerraFacet {
    fn facet_id(&self) -> FacetId { self.meta.id }
    fn domain_id(&self) -> DomainId { self.meta.domain_id }

    fn domain_surface(&self) -> DomainSurface {
        DomainSurface {
            id: self.meta.domain_id,
            facet_id: self.meta.id,
            reachable: self.reachable,
        }
    }
}
