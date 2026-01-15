// src/spine/lattice.rs
//
// Spine Lattice = intrinsic geometry of the layers.
// Defines:
// - structural_basis: what a layer stands on (non-adjacent grounding for life layers)
// - acts_forward: what a layer lifts into next (ordered progression)
//
// No actors, no seats, no domains, no policy.

use super::SpineLayer;

#[derive(Debug, Clone, Copy)]
pub struct SpineLattice {
    pub layer: SpineLayer,
    pub structural_basis: Option<SpineLayer>,
    pub acts_forward: Option<SpineLayer>,
}

impl SpineLayer {
    pub fn lattice(self) -> SpineLattice {
        match self {
            SpineLayer::Nullus => SpineLattice {
                layer: self,
                structural_basis: None,
                acts_forward: Some(SpineLayer::Aether),
            },
            SpineLayer::Aether => SpineLattice {
                layer: self,
                structural_basis: None,
                acts_forward: Some(SpineLayer::Terra),
            },
            SpineLayer::Terra => SpineLattice {
                layer: self,
                structural_basis: None,
                acts_forward: Some(SpineLayer::Numen),
            },
            SpineLayer::Numen => SpineLattice {
                layer: self,
                structural_basis: Some(SpineLayer::Terra),
                acts_forward: Some(SpineLayer::Lima),
            },
            SpineLayer::Lima => SpineLattice {
                layer: self,
                structural_basis: Some(SpineLayer::Aether),
                acts_forward: Some(SpineLayer::Corporeal),
            },
            SpineLayer::Corporeal => SpineLattice {
                layer: self,
                structural_basis: Some(SpineLayer::Nullus),
                acts_forward: None,
            },
        }
    }

    /// True if this layer requires a non-adjacent structural grounding.
    pub fn requires_structural_basis(self) -> bool {
        self.lattice().structural_basis.is_some()
    }

    /// True if this layer is one of the "life" layers (mind-ish).
    pub fn is_life_layer(self) -> bool {
        matches!(self, SpineLayer::Numen | SpineLayer::Lima | SpineLayer::Corporeal)
    }

    pub fn is_physics_layer(self) -> bool {
        !self.is_life_layer()
    }

    /// True only for adjacent forward progression (no skipping).
    pub fn can_act_forward_to(self, next: SpineLayer) -> bool {
        self.lattice().acts_forward == Some(next)
    }
}

#[derive(Debug, Clone)]
pub enum SpineError {
    MissingStructuralBasis { layer: SpineLayer, required: SpineLayer },
    MissingLayer { layer: SpineLayer },
}

/// Validates that every included layer has its required structural basis also present.
/// Order does not matter; this checks set completeness for grounding.
pub fn validate_lattice_stack(layers: &[SpineLayer]) -> Result<(), SpineError> {
    for &layer in layers {
        if let Some(required) = layer.lattice().structural_basis {
            if !layers.contains(&required) {
                return Err(SpineError::MissingStructuralBasis { layer, required });
            }
        }
    }
    Ok(())
}

/// Convenience: validates that all 6 layers are present (a "full cube").
pub fn validate_full_spine(layers: &[SpineLayer]) -> Result<(), SpineError> {
    for &layer in &SpineLayer::ALL {
        if !layers.contains(&layer) {
            return Err(SpineError::MissingLayer { layer });
        }
    }
    validate_lattice_stack(layers)
}
