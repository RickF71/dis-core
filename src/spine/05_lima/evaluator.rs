// src/spine/05_lima/evaluator.rs

use crate::spine::lima::{
    Transition,
    LimaContext,
    Decision,
    DenyReason,
};

use crate::spine::SpineLayer;

use super::receipt::TransitionReceipt;

pub struct LimaEvaluator;

impl LimaEvaluator {
    pub fn evaluate(t: &Transition, ctx: &LimaContext) -> Decision {
        // --- 0) Domain sanity
        if t.domain != ctx.domain {
            return Decision::deny(DenyReason::Other {
                code: "deny:lima:domain_mismatch".into(),
                detail: "transition domain != context domain".into(),
            });
        }

        // --- 1) Freeze ribbon
        if ctx.freeze.is_frozen() {
            let scope = ctx
                .freeze
                .active_scopes
                .iter()
                .next()
                .cloned()
                .unwrap_or_else(|| crate::spine::lima::types::FreezeScopeId("unknown".into()));

            return Decision::deny(DenyReason::Frozen { scope });
        }

        // --- 2) Aether: seat binding
        match ctx.aether.occupant(&t.seat) {
            None => return Decision::deny(DenyReason::SeatNotBound { seat: t.seat.clone() }),
            Some(occ) if *occ != t.actor => {
                return Decision::deny(DenyReason::ActorNotOccupant {
                    seat: t.seat.clone(),
                    actor: t.actor.clone(),
                })
            }
            Some(_) => {}
        }

        // --- 3) Terra
        if !ctx.terra.has_path(&t.from, &t.to) {
            return Decision::deny(DenyReason::NoStructuralPath {
                from: t.from.clone(),
                to: t.to.clone(),
            });
        }

        // --- 4) Numen
        if !ctx.numen.meaning_applies(t) {
            return Decision::deny(DenyReason::MeaningNotApplicable {
                meaning: t.meaning.clone(),
            });
        }

        // --- 5) Contracts
        for c in ctx.numen.contracts_for_meaning(&t.meaning) {
            if !ctx.numen.contract_satisfied(&c, t) {
                return Decision::deny(DenyReason::ContractNotSatisfied { contract: c });
            }
        }

        // --- 6) Turn
        if let Some(active) = &ctx.turn.active_seat {
            if *active != t.seat {
                return Decision::deny(DenyReason::NotYourTurn { seat: t.seat.clone() });
            }
        }

        // --- 7) Domain rules
        if let Some(rules) = ctx.domain_rules {
            if let Some(reason) = rules.additional_checks(t) {
                return Decision::deny(reason);
            }
        }

        Decision::Allow
    }
}

// ------------------------------------------------------------
// Receipt emission (PURE WRAPPER — no authority change)
// ------------------------------------------------------------

pub fn evaluate_with_receipt(
    t: &Transition,
    ctx: &LimaContext,
) -> TransitionReceipt {
    let decision = LimaEvaluator::evaluate(t, ctx);

    TransitionReceipt {
        domain: t.domain.clone(),
        actor: t.actor.clone(),
        seat: t.seat.clone(),
        from: t.from.clone(),
        to: t.to.clone(),
        meaning: t.meaning.clone(),
        tick: t.tick,
        layer: SpineLayer::Lima,
        decision,
    }
}
