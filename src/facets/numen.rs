use crate::id::DomainId;
use super::{FacetMeta, FacetKind, FacetId};
use super::interfaces::*;
use super::terra::TerraFacet;

#[derive(Debug)]
pub struct NumenFacet {
    pub meta: FacetMeta,
}

impl NumenFacet {
    pub fn mint_from_parent(parent: &TerraFacet) -> Self {
        Self {
            meta: FacetMeta::new(parent.meta.domain_id, FacetKind::Numen, Some(parent.meta.id)),
        }
    }

    pub fn mint_lima(&self) -> super::lima::LimaFacet {
        super::lima::LimaFacet::mint_from_parent(self)
    }
}

impl ResolutionIfc for NumenFacet {
    fn facet_id(&self) -> FacetId { self.meta.id }
    fn domain_id(&self) -> DomainId { self.meta.domain_id }

    fn resolve(&mut self, _req: ResolveRequest) -> ResolveResult {
        // stub: Numen can be deferred; keep interface shape stable
        ResolveResult { decided: false }
    }
}
