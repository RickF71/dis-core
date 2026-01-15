use crate::id::DomainId;
use super::{nullus::NullusFacet, aether::AetherFacet, terra::TerraFacet, numen::NumenFacet, lima::LimaFacet, corporeal::CorporealFacet};

/// Domain prototype: a handle to the 6-facet mint chain.
/// This is the "domain as a whole" prototype you asked for.
#[derive(Debug)]
pub struct DomainStack {
    pub domain_id: DomainId,

    pub nullus: NullusFacet,
    pub aether: AetherFacet,
    pub terra: TerraFacet,
    pub numen: NumenFacet,
    pub lima: LimaFacet,
    pub corporeal: CorporealFacet,
}

impl DomainStack {
    /// Mint an entire domain facet chain deterministically.
    /// No threads, no sockets, no policy.
    pub fn mint(domain_id: DomainId) -> Self {
        let nullus = NullusFacet::mint_root(domain_id);
        let aether = nullus.mint_aether();
        let terra = aether.mint_terra();
        let numen = terra.mint_numen();
        let lima = numen.mint_lima();
        let corporeal = lima.mint_corporeal();

        Self {
            domain_id,
            nullus,
            aether,
            terra,
            numen,
            lima,
            corporeal,
        }
    }
}
