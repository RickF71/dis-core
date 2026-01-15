use uuid::Uuid;
use crate::id::DomainId;
use super::FacetKind;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct FacetId(pub Uuid);

#[derive(Debug, Clone)]
pub struct FacetMeta {
    pub id: FacetId,
    pub domain_id: DomainId,
    pub kind: FacetKind,

    /// Immediate parent facet in the *same* domain chain.
    /// (Nullus has none; Corporeal parent is Lima, etc.)
    pub parent_facet: Option<FacetId>,
}

impl FacetMeta {
    pub fn new(domain_id: DomainId, kind: FacetKind, parent_facet: Option<FacetId>) -> Self {
        Self {
            id: FacetId(Uuid::new_v4()),
            domain_id,
            kind,
            parent_facet,
        }
    }
}
