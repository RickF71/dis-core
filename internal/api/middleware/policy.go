package middleware

import (
	"context"
	"net/http"

	"dis-core/internal/policy"
)

const policyContextKey contextKey = "policy_engine"

// WithPolicyEngine middleware injects the policy engine into the request context
func WithPolicyEngine(engine policy.PolicyEngine) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Inject the policy engine into the request context
			ctx := context.WithValue(r.Context(), policyContextKey, engine)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromPolicyEngine retrieves the policy engine from the request context
func FromPolicyEngine(ctx context.Context) policy.PolicyEngine {
	if engine, ok := ctx.Value(policyContextKey).(policy.PolicyEngine); ok {
		return engine
	}
	return nil
}
