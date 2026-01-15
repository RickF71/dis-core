use crate::id::DomainId;
use super::{FacetMeta, FacetKind, FacetId};
use super::interfaces::*;
use super::numen::NumenFacet;

#[derive(Debug)]
pub struct LimaFacet {
    pub meta: FacetMeta,
    notes: Vec<Note>,
}

impl LimaFacet {
    pub fn mint_from_parent(parent: &NumenFacet) -> Self {
        Self {
            meta: FacetMeta::new(parent.meta.domain_id, FacetKind::Lima, Some(parent.meta.id)),
            notes: vec![],
        }
    }

    pub fn mint_corporeal(&self) -> super::corporeal::CorporealFacet {
        super::corporeal::CorporealFacet::mint_from_parent(self)
    }
}

impl MeaningIfc for LimaFacet {
    fn facet_id(&self) -> FacetId { self.meta.id }
    fn domain_id(&self) -> DomainId { self.meta.domain_id }

    fn annotate(&mut self, note: Note) {
        self.notes.push(note);
    }

    fn meaning_surface(&self) -> MeaningSurface {
        MeaningSurface {
            facet_id: self.meta.id,
            notes_count: self.notes.len(),
        }
    }
}
