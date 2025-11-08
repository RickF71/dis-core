package policy

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/rego"
)

type OPAEngine struct {
	gatesRego  *rego.PreparedEvalQuery
	riskRego   *rego.PreparedEvalQuery
	freezeRego *rego.PreparedEvalQuery
}

var _ PolicyEngine = (*OPAEngine)(nil)

// PolicyEngine is the interface for runtime policy evaluation.
type PolicyEngine interface {
	EvaluateAction(input map[string]interface{}) (*PolicyDecision, error)
}

// NewEngine creates a new OPA-based policy engine with the given modules.
func NewEngine(modules map[string]string) (*OPAEngine, error) {
	ctx := context.Background()

	// --- Gates (allow/deny logic)
	gq, err := rego.New(
		rego.Query("data.gates.allow"),
		rego.Module("gates.rego", modules["gates.rego"]),
	).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("prepare gates: %w", err)
	}

	// --- Risk (scoring or threshold logic)
	rq, err := rego.New(
		rego.Query("data.risk.score"),
		rego.Module("risk.rego", modules["risk.rego"]),
	).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("prepare risk: %w", err)
	}

	// --- Freeze (domain freeze / unfreeze / override)
	fq, err := rego.New(
		rego.Query("data.freeze.active"),
		rego.Module("freeze.rego", modules["freeze.rego"]),
	).PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("prepare freeze: %w", err)
	}

	return &OPAEngine{
		gatesRego:  &gq,
		riskRego:   &rq,
		freezeRego: &fq,
	}, nil
}

func (e *OPAEngine) EvaluateAction(input map[string]interface{}) (*PolicyDecision, error) {
	if e.gatesRego == nil && e.riskRego == nil && e.freezeRego == nil {
		return &PolicyDecision{
			Allow:  true,
			Reason: "no policies loaded; default allow",
		}, nil
	}

	ctx := context.Background()
	allow := true
	details := map[string]interface{}{}

	// --- GATES
	if e.gatesRego != nil {
		rs, err := e.gatesRego.Eval(ctx, rego.EvalInput(input))
		if err != nil {
			return nil, fmt.Errorf("gates eval: %w", err)
		}
		if len(rs) > 0 && len(rs[0].Expressions) > 0 {
			if b, ok := rs[0].Expressions[0].Value.(bool); ok {
				details["gates.allow"] = b
				if !b {
					allow = false
				}
			}
		}
	}

	// --- RISK
	if e.riskRego != nil {
		rs, err := e.riskRego.Eval(ctx, rego.EvalInput(input))
		if err != nil {
			return nil, fmt.Errorf("risk eval: %w", err)
		}
		if len(rs) > 0 && len(rs[0].Expressions) > 0 {
			if f, ok := rs[0].Expressions[0].Value.(float64); ok {
				details["risk.score"] = f
			}
		}
	}

	// --- FREEZE
	if e.freezeRego != nil {
		rs, err := e.freezeRego.Eval(ctx, rego.EvalInput(input))
		if err != nil {
			return nil, fmt.Errorf("freeze eval: %w", err)
		}
		if len(rs) > 0 && len(rs[0].Expressions) > 0 {
			if frozen, ok := rs[0].Expressions[0].Value.(bool); ok {
				details["freeze.active"] = frozen
				if frozen {
					allow = false
				}
			}
		}
	}

	return &PolicyDecision{
		Allow:   allow,
		Reason:  "evaluated via gates/risk/freeze",
		Details: details,
	}, nil
}
