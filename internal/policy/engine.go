package policy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/open-policy-agent/opa/rego"
)

type OPAEngine struct {
	gatesRego *rego.PreparedEvalQuery
	riskRego  *rego.PreparedEvalQuery
}

func NewOPAEngine() (*OPAEngine, error) {
	base := "./policies" // or os.Getenv("DIS_POLICY_PATH")

	gates, err := os.ReadFile(filepath.Join(base, "gates.rego"))
	if err != nil {
		return nil, fmt.Errorf("read gates.rego: %w", err)
	}

	risk, err := os.ReadFile(filepath.Join(base, "risk.rego"))
	if err != nil {
		return nil, fmt.Errorf("read risk.rego: %w", err)
	}

	gq, err := rego.New(
		rego.Query("data.gates.allow"),
		rego.Module("gates.rego", string(gates)),
	).PrepareForEval(context.Background())
	if err != nil {
		return nil, fmt.Errorf("prepare gates: %w", err)
	}

	rq, err := rego.New(
		rego.Query("data.risk.score"),
		rego.Module("risk.rego", string(risk)),
	).PrepareForEval(context.Background())
	if err != nil {
		return nil, fmt.Errorf("prepare risk: %w", err)
	}

	return &OPAEngine{gatesRego: &gq, riskRego: &rq}, nil
}

func (e *OPAEngine) EvaluateAction(input map[string]interface{}) (*PolicyDecision, error) {
	ctx := context.Background()
	gatesRes, err := e.gatesRego.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return nil, fmt.Errorf("gates eval: %w", err)
	}
	allow := false
	if len(gatesRes) > 0 {
		allow, _ = gatesRes[0].Expressions[0].Value.(bool)
	}
	riskRes, err := e.riskRego.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return nil, fmt.Errorf("risk eval: %w", err)
	}
	risk := 0.0
	if len(riskRes) > 0 {
		if v, ok := riskRes[0].Expressions[0].Value.(float64); ok {
			risk = v
		}
	}
	return &PolicyDecision{
		Allow:     allow,
		RiskScore: risk,
	}, nil
}
