package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"dis-core/internal/auth"
	"dis-core/internal/contextx"
	"dis-core/internal/policy"
)

// WithPolicyEnforcement evaluates policy for mutating requests and blocks
// denied actions. It also attaches the PolicyDecision to the request context
// so handlers can include decision metadata in receipts.
func WithPolicyEnforcement() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip evaluation for safe methods
			if r.Method == http.MethodGet || r.Method == http.MethodOptions || r.Method == http.MethodHead {
				next.ServeHTTP(w, r)
				return
			}

			// Obtain policy engine from context
			engine := FromPolicyEngine(r.Context())
			if engine == nil {
				// Nothing to enforce
				next.ServeHTTP(w, r)
				return
			}

			// Build a compact input for evaluation. Include method, path,
			// provenance (if any), and actor identity where available.
			input := map[string]interface{}{
				"http_method": r.Method,
				"path":        r.URL.Path,
			}

			// Attach actor info from ActiveUser when available
			if au := auth.GetActiveUser(r); au != nil {
				if au.IsBound() {
					input["actor_corporeal_uid"] = au.CorporealDomainUID
					input["actor_corporeal_id"] = au.CorporealDomainID
				} else if au.HasExternalUID {
					input["actor_external_uid"] = au.ExternalUID
				}
			}

			// Prefer chi path param "id" when present (common domain endpoints)
			if id := chi.URLParam(r, "id"); id != "" {
				input["domain_id"] = id
			}

			// Evaluate using the policy engine
			decision, err := engine.EvaluateAction(r.Context(), input)
			if err != nil {
				log.Printf("policy evaluate error: %v", err)
				http.Error(w, "policy evaluation failed", http.StatusInternalServerError)
				return
			}
			if decision == nil {
				// No decision -> default allow
				next.ServeHTTP(w, r)
				return
			}

			// If denied, return a 403 with reason
			if !decision.Allow {
				http.Error(w, "action denied by policy: "+decision.Reason, http.StatusForbidden)
				return
			}

			// Attach decision to context for downstream handlers (receipts, logging)
			// Store both the typed PolicyDecision (for API-level consumers) and
			// a generic map form (for low-level core packages that cannot
			// import the policy package to avoid import cycles).
			ctx := policy.WithPolicyDecision(r.Context(), decision)
			// build a lightweight map representation
			dm := map[string]interface{}{
				"allow":  decision.Allow,
				"reason": decision.Reason,
			}
			if !decision.Timestamp.IsZero() {
				dm["timestamp"] = decision.Timestamp.Format(time.RFC3339)
			}
			if len(decision.Roles) > 0 {
				dm["roles"] = decision.Roles
			}
			if decision.Details != nil {
				for k, v := range decision.Details {
					dm[k] = v
				}
			}
			ctx = contextx.WithPolicyDecisionMap(ctx, dm)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromPolicyDecision retrieves the PolicyDecision attached by the enforcement middleware.
func FromPolicyDecision(ctx context.Context) (*policy.PolicyDecision, bool) {
	return policy.FromContext(ctx)
}
