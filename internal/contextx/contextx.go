package contextx

import "context"

// small helpers to store generic values on context without importing higher-level
// packages. These intentionally avoid typed dependencies so low-level core
// packages (authority, identity) can consume provenance information without
// creating import cycles.

type key string

const (
	policyDecisionKey key = "policyDecision"
	originDomainKey   key = "originDomain"
)

// WithPolicyDecisionMap stores a generic map representation of a policy decision.
func WithPolicyDecisionMap(ctx context.Context, m map[string]interface{}) context.Context {
	return context.WithValue(ctx, policyDecisionKey, m)
}

// PolicyDecisionMapFromContext retrieves a generic map representation of a
// policy decision if present.
func PolicyDecisionMapFromContext(ctx context.Context) (map[string]interface{}, bool) {
	if ctx == nil {
		return nil, false
	}
	if v := ctx.Value(policyDecisionKey); v != nil {
		if m, ok := v.(map[string]interface{}); ok {
			return m, true
		}
	}
	return nil, false
}

// WithOriginDomain stores a *domain.Domain on the context to indicate the
// origin domain for subsequent receipt emission and provenance tracking.
func WithOriginDomain(ctx context.Context, d interface{}) context.Context {
	return context.WithValue(ctx, originDomainKey, d)
}

// OriginDomainFromContext retrieves the origin domain if present. The
// value is returned as interface{} to avoid importing domain package here;
// callers should type-assert to the concrete *domain.Domain when needed.
func OriginDomainFromContext(ctx context.Context) (interface{}, bool) {
	if ctx == nil {
		return nil, false
	}
	if v := ctx.Value(originDomainKey); v != nil {
		return v, true
	}
	return nil, false
}
