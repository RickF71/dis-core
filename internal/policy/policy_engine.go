package policy

// PolicyEngine is a light wrapper that exposes EvaluateAction
// through the existing OPAEngine implementation.

// PolicyEngineImpl is a light wrapper that exposes EvaluateAction
// through the existing OPAEngine implementation.
type PolicyEngineImpl struct {
	engine *OPAEngine
}

// NewPolicyEngineImpl wraps an OPAEngine so other packages
// can depend on this higher-level name instead of OPA details.
func NewPolicyEngineImpl(e *OPAEngine) *PolicyEngineImpl {
	return &PolicyEngineImpl{engine: e}
}

// EvaluateAction proxies to the OPAEngine's method.
func (p *PolicyEngineImpl) EvaluateAction(input map[string]interface{}) (*PolicyDecision, error) {
	if p.engine == nil {
		return nil, nil
	}
	return p.engine.EvaluateAction(input)
}
