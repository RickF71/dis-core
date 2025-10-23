package policy

import (
	"context"
	"fmt"
	"os"

	"github.com/open-policy-agent/opa/rego"
)

// NewEngine builds an OPAEngine from the file paths in cfg.
// (Freeze/Cedar/thresholds are ignored here for the minimal bring-up.)
func NewEngine(cfg EngineConfig) (*OPAEngine, error) {
	gatesSrc, err := os.ReadFile(cfg.PathGatesRego)
	if err != nil {
		return nil, fmt.Errorf("read gates.rego: %w", err)
	}
	riskSrc, err := os.ReadFile(cfg.PathRiskRego)
	if err != nil {
		return nil, fmt.Errorf("read risk.rego: %w", err)
	}

	gq, err := rego.New(
		rego.Query("data.gates.allow"),
		rego.Module("gates.rego", string(gatesSrc)),
	).PrepareForEval(context.Background())
	if err != nil {
		return nil, fmt.Errorf("prepare gates: %w", err)
	}

	rq, err := rego.New(
		rego.Query("data.risk.score"),
		rego.Module("risk.rego", string(riskSrc)),
	).PrepareForEval(context.Background())
	if err != nil {
		return nil, fmt.Errorf("prepare risk: %w", err)
	}

	return &OPAEngine{gatesRego: &gq, riskRego: &rq}, nil
}
