package api

import (
	"net/http"

	"dis-core/internal/auth"
)

// WithTestActiveUser attaches a bound ActiveUser to the request context for tests.
// domainUID should be a string UUID representing the corporeal domain UID.
func WithTestActiveUser(r *http.Request, domainUID string) *http.Request {
	au := &auth.ActiveUser{
		CorporealDomainUID: domainUID,
		Bound:              true,
		CorporealDomainID:  1,
	}
	ctx := auth.WithActiveUser(r.Context(), au)
	return r.WithContext(ctx)
}

// NewTestActiveUser returns a test ActiveUser instance (not used by HTTP helpers).
func NewTestActiveUser(domainUID string) *auth.ActiveUser {
	return &auth.ActiveUser{
		CorporealDomainUID: domainUID,
		Bound:              true,
		CorporealDomainID:  1,
	}
}
