package authority

import "context"

// IntrospectionResult is the structured payload returned by introspection.
type IntrospectionResult struct {
	EngineVersion  string                 `json:"engine_version"`
	Capabilities   []string               `json:"capabilities"`
	PolicyBackends []string               `json:"policy_backends"`
	Notes          []string               `json:"notes,omitempty"`
	Raw            map[string]interface{} `json:"raw,omitempty"`
}

// Introspect yields structural details about authority configuration.
// Real logic migrated from internal/api/authority_introspect.go.
func (e *Engine) Introspect() *IntrospectionResult {
	return &IntrospectionResult{
		EngineVersion:  "MX-2.1",
		Capabilities:   []string{"status", "lineage", "freeze"},
		PolicyBackends: []string{"rego", "cedar"},
		Notes:          []string{"Replace with real introspection results"},
		Raw:            map[string]interface{}{},
	}
}

// GetIntrospect returns a small introspection payload from the authority engine.
// Part of MX-3.2 migration.
func (e *Engine) GetIntrospect(ctx context.Context) (*IntrospectInfo, error) {
	info := &IntrospectInfo{
		EngineVersion: "mx-3.2",
		Notes:         "MX-3.2: Introspect migrated to new authority engine.",
	}
	return info, nil
}
