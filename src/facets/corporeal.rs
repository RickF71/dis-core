use crate::id::DomainId;
use super::{FacetMeta, FacetKind, FacetId};
use super::interfaces::*;
use super::lima::LimaFacet;

#[derive(Debug)]
pub struct CorporealFacet {
    pub meta: FacetMeta,
    next_session: u64,
}

impl CorporealFacet {
    pub fn mint_from_parent(parent: &LimaFacet) -> Self {
        Self {
            meta: FacetMeta::new(parent.meta.domain_id, FacetKind::Corporeal, Some(parent.meta.id)),
            next_session: 1,
        }
    }
}

impl InteractionIfc for CorporealFacet {
    fn facet_id(&self) -> FacetId { self.meta.id }
    fn domain_id(&self) -> DomainId { self.meta.domain_id }

    fn open_session(&mut self) -> SessionId {
        let sid = SessionId(self.next_session);
        self.next_session += 1;
        sid
    }

    fn accept_input(&mut self, _sid: SessionId, _input: InputEvent) -> () {
        // stub: interaction intake only
    }

    fn render(&self, _sid: SessionId) -> RenderFrame {
        RenderFrame {
            ok: true,
            message: "corporeal.render.stub".to_string(),
        }
    }

    fn mint_child_domain(&self) -> DomainId {
        // stub: real implementation should mint a new DomainId and emit lineage receipts
        DomainId(uuid::Uuid::new_v4())
    }
}
