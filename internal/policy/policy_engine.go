package policy

// PolicyEngine is a light wrapper that exposes EvaluateAction
// through the existing OPAEngine implementation.
type PolicyEngine struct {
	engine *OPAEngine
}

// NewPolicyEngine wraps an OPAEngine so other packages
// can depend on this higher-level name instead of OPA details.
func NewPolicyEngine(e *OPAEngine) *PolicyEngine {
	return &PolicyEngine{engine: e}
}

// EvaluateAction proxies to the OPAEngine's method.
func (p *PolicyEngine) EvaluateAction(input map[string]interface{}) (*PolicyDecision, error) {
	if p.engine == nil {
		return nil, nil
	}
	return p.engine.EvaluateAction(input)
}
