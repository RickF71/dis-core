// src/spine/05_lima/context.rs
use std::collections::{HashMap, HashSet};

use crate::spine::lima::types::*;
use crate::spine::lima::transition::Transition;

//
// Views (interfaces) that Lima evaluates against.
// These keep Lima orthogonal: it depends on *contracts* of information,
// not concrete storage or domain implementations.
//

pub trait TerraView {
    /// True if `from -> to` is structurally reachable (however your Terra defines it).
    /// Minimal baseline: direct adjacency.
    fn has_edge(&self, from: &TerraNodeId, to: &TerraNodeId) -> bool;

    /// Optional: allow domains to model richer reachability (paths).
    fn has_path(&self, from: &TerraNodeId, to: &TerraNodeId) -> bool {
        self.has_edge(from, to)
    }
}

pub trait NumenView {
    /// True if the meaning is applicable to this transition.
    /// Example: “diagonal” meaning only applies when move is diagonal.
    fn meaning_applies(&self, t: &Transition) -> bool;

    /// Contracts that are associated with this meaning (if you want Lima to check them).
    /// If you don’t need this yet, return empty.
    fn contracts_for_meaning(&self, _meaning: &MeaningId) -> HashSet<ContractId> {
        HashSet::new()
    }

    /// Optional: contract precondition check. Domains can implement.
    fn contract_satisfied(&self, _contract: &ContractId, _t: &Transition) -> bool {
        true
    }
}

/// Domain-specific constraint hook (optional).
/// Keep it small: it should *refine* decisions, not replace Lima.
pub trait DomainRules {
    /// Called after baseline checks pass. Return Some(DenyReason) to deny.
    fn additional_checks(&self, _t: &Transition) -> Option<crate::spine::lima::reasons::DenyReason> {
        None
    }
}

#[derive(Debug, Clone)]
pub struct AetherBindings {
    /// seat_id -> current occupant actor_id
    pub seat_occupant: HashMap<SeatId, ActorId>,
}

impl AetherBindings {
    pub fn new() -> Self {
        Self { seat_occupant: HashMap::new() }
    }

    pub fn bind(&mut self, seat: SeatId, actor: ActorId) {
        self.seat_occupant.insert(seat, actor);
    }

    pub fn unbind(&mut self, seat: &SeatId) {
        self.seat_occupant.remove(seat);
    }

    pub fn occupant(&self, seat: &SeatId) -> Option<&ActorId> {
        self.seat_occupant.get(seat)
    }
}

/// Minimal turn state: which seat currently has the move (domain-defined).
#[derive(Debug, Clone)]
pub struct TurnState {
    pub active_seat: Option<SeatId>,
}

impl TurnState {
    pub fn none() -> Self { Self { active_seat: None } }
}

/// Freeze is modeled as a set of active deny scopes.
/// (You can later attach TTL, break-glass receipts, etc.)
#[derive(Debug, Clone)]
pub struct FreezeState {
    pub active_scopes: HashSet<FreezeScopeId>,
}

impl FreezeState {
    pub fn none() -> Self { Self { active_scopes: HashSet::new() } }

    pub fn is_frozen(&self) -> bool {
        !self.active_scopes.is_empty()
    }
}

pub struct LimaContext<'a> {
    pub domain: DomainId,

    pub aether: &'a AetherBindings,
    pub terra: &'a dyn TerraView,
    pub numen: &'a dyn NumenView,

    pub turn: TurnState,
    pub freeze: FreezeState,

    /// Optional domain-specific rules.
    pub domain_rules: Option<&'a dyn DomainRules>,
}

impl<'a> LimaContext<'a> {
    pub fn new(
        domain: DomainId,
        aether: &'a AetherBindings,
        terra: &'a dyn TerraView,
        numen: &'a dyn NumenView,
    ) -> Self {
        Self {
            domain,
            aether,
            terra,
            numen,
            turn: TurnState::none(),
            freeze: FreezeState::none(),
            domain_rules: None,
        }
    }
}
