package policy

import (
	"context"
	"time"
)

// PolicyDecision is the result of a policy evaluation.
type PolicyDecision struct {
	Allow      bool                   `json:"allow"`
	Reason     string                 `json:"reason,omitempty"`
	BreakGlass bool                   `json:"break_glass,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Roles      []string               `json:"roles,omitempty"`
}

// context key used to store PolicyDecision on a context.Context. Exported
// via helper functions below so other packages don't need to depend on the
// middleware package's key type.
type contextKey string

const policyDecisionContextKey contextKey = "policyDecision"

// WithPolicyDecision returns a new context containing the given PolicyDecision.
func WithPolicyDecision(ctx context.Context, d *PolicyDecision) context.Context {
	return context.WithValue(ctx, policyDecisionContextKey, d)
}

// FromContext attempts to retrieve a PolicyDecision from a context.Context.
func FromContext(ctx context.Context) (*PolicyDecision, bool) {
	if ctx == nil {
		return nil, false
	}
	if v := ctx.Value(policyDecisionContextKey); v != nil {
		if pd, ok := v.(*PolicyDecision); ok {
			return pd, true
		}
	}
	return nil, false
}
