package policy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/open-policy-agent/opa/rego"
)

// -------------------------------------------------------------
// Engine Definition
// -------------------------------------------------------------

type OPAEngine struct {
	gatesRego  *rego.PreparedEvalQuery
	riskRego   *rego.PreparedEvalQuery
	freezeRego *rego.PreparedEvalQuery
}

var _ PolicyEngine = (*OPAEngine)(nil)

// PolicyEngine is the interface for runtime policy evaluation.
type PolicyEngine interface {
	EvaluateAction(ctx context.Context, input map[string]interface{}) (*PolicyDecision, error)
}

// -------------------------------------------------------------
// Module Loader
// -------------------------------------------------------------

// LoadLocalModules reads local .rego files under internal/policy/.
func LoadLocalModules() (map[string]string, error) {
	base := "internal/policy"
	files := []string{"gates.rego", "risk.rego", "freeze.rego"}
	modules := map[string]string{}

	for _, f := range files {
		path := filepath.Join(base, f)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		modules[f] = string(data)
	}

	return modules, nil
}

// -------------------------------------------------------------
// Engine Construction
// -------------------------------------------------------------

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

// -------------------------------------------------------------
// Helpers
// -------------------------------------------------------------

// evalDetails executes a Rego query returning an object and merges keys into details.
func evalDetails(ctx context.Context, moduleName, query string, moduleSrc string, input map[string]interface{}, prefix string, details map[string]interface{}) {
	q, err := rego.New(
		rego.Query(query),
		rego.Module(moduleName, moduleSrc),
	).PrepareForEval(ctx)
	if err != nil {
		return
	}

	rs, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil || len(rs) == 0 || len(rs[0].Expressions) == 0 {
		return
	}

	if m, ok := rs[0].Expressions[0].Value.(map[string]interface{}); ok {
		for k, v := range m {
			details[fmt.Sprintf("%s.%s", prefix, k)] = v
		}
	}
}

// -------------------------------------------------------------
// Core Evaluation
// -------------------------------------------------------------

func (e *OPAEngine) EvaluateAction(ctx context.Context, input map[string]interface{}) (*PolicyDecision, error) {
	if e.gatesRego == nil && e.riskRego == nil && e.freezeRego == nil {
		return &PolicyDecision{
			Allow:     true,
			Reason:    "no policies loaded; default allow",
			Timestamp: time.Now().UTC(),
		}, nil
	}

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
		// Merge any data.gates.details
		evalDetails(ctx, "gates.rego", "data.gates.details", e.getModule("gates.rego"), input, "gates", details)
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
		evalDetails(ctx, "risk.rego", "data.risk.details", e.getModule("risk.rego"), input, "risk", details)
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
		evalDetails(ctx, "freeze.rego", "data.freeze.details", e.getModule("freeze.rego"), input, "freeze", details)
	}

	return &PolicyDecision{
		Allow:     allow,
		Reason:    "evaluated via gates/risk/freeze",
		Details:   details,
		Timestamp: time.Now().UTC(),
	}, nil
}

// EvaluateWithSeatPolicies evaluates action with per-seat REGO policies added
// Phase S5: Per-seat policy integration
func (e *OPAEngine) EvaluateWithSeatPolicies(ctx context.Context, input map[string]interface{}, seatPolicies map[string]string) (*PolicyDecision, error) {
	// First run base evaluation (gates/risk/freeze)
	baseDecision, err := e.EvaluateAction(ctx, input)
	if err != nil {
		return nil, err
	}

	// If base policies deny, per-seat policies cannot override
	if !baseDecision.Allow {
		return baseDecision, nil
	}

	// Evaluate all per-seat policies
	// Each seat policy can tighten (deny) but not loosen restrictions
	seatDetails := map[string]interface{}{}
	seatAllow := true

	for seatID, regoText := range seatPolicies {
		// Extract package name from REGO to build correct query path
		// Default to generic path if we can't extract
		queryPath := "data.dis.seat.export_allow"

		// Compile and evaluate this seat's policy
		query := rego.New(
			rego.Query(queryPath),
			rego.Module(fmt.Sprintf("seat_%s.rego", seatID), regoText),
		)

		prepared, err := query.PrepareForEval(ctx)
		if err != nil {
			// Log error but don't fail - skip invalid policies
			seatDetails[fmt.Sprintf("seat.%s.error", seatID)] = err.Error()
			continue
		}

		rs, err := prepared.Eval(ctx, rego.EvalInput(input))
		if err != nil {
			seatDetails[fmt.Sprintf("seat.%s.error", seatID)] = err.Error()
			continue
		}

		// Check export_allow result - if undefined (no expressions), treat as deny
		seatPolicyAllow := false
		if len(rs) > 0 && len(rs[0].Expressions) > 0 {
			if allow, ok := rs[0].Expressions[0].Value.(bool); ok {
				seatPolicyAllow = allow
			}
		}

		seatDetails[fmt.Sprintf("seat.%s.allow", seatID)] = seatPolicyAllow
		if !seatPolicyAllow {
			seatAllow = false
		}
	}

	// Merge seat details into base decision
	for k, v := range seatDetails {
		baseDecision.Details[k] = v
	}

	// If any seat policy denies, overall decision is deny
	if !seatAllow {
		baseDecision.Allow = false
		baseDecision.Reason = "denied by per-seat policy"
	} else {
		baseDecision.Reason = fmt.Sprintf("evaluated via gates/risk/freeze + %d seat policies", len(seatPolicies))
	}

	return baseDecision, nil
}

// Reload rebuilds the OPAEngine with domain-specific modules
func (e *OPAEngine) Reload(ctx context.Context, domainID string) error {
	// Load domain-specific modules with fallbacks
	modules, err := LoadDomainModules(domainID)
	if err != nil {
		return fmt.Errorf("load domain modules for %s: %w", domainID, err)
	}

	// Create a new engine with the loaded modules
	newEngine, err := NewEngine(modules)
	if err != nil {
		return fmt.Errorf("create new engine: %w", err)
	}

	// Replace current engine state with the new one
	e.gatesRego = newEngine.gatesRego
	e.riskRego = newEngine.riskRego
	e.freezeRego = newEngine.freezeRego

	return nil
}

// getModule retrieves the source code of a module from disk for detail evaluation.
func (e *OPAEngine) getModule(name string) string {
	path := filepath.Join("internal", "policy", name)
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
