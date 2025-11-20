package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ExternalAuthMiddleware reads the X-External-User header OR browser session
// and attaches an ActiveUser to the request context.
// Supports both:
//   - X-External-User header (dev mode, explicit user ID)
//   - dis_browser_session cookie (QR auth flow, session-based)
func ExternalAuthMiddleware(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try X-External-User header first (dev mode)
			externalUID := strings.TrimSpace(r.Header.Get("X-External-User"))

			// (Legacy QR/cookie flow removed)

			// If no external UID was provided, do not attach an ActiveUser.
			// This keeps anonymous requests unauthenticated so handlers that
			// require authentication can return 401.
			if externalUID == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Create ActiveUser for requests that provide an external UID
			user := &ActiveUser{
				ExternalUID:    externalUID,
				HasExternalUID: true,
				Bound:          false, // Will be set by resolver middleware
			}

			// Attach to context
			ctx := WithActiveUser(r.Context(), user)

			// Continue with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ActiveUserResolverMiddleware resolves the external UID to a corporeal domain
// if a binding exists in identity_bindings table.
// Must run AFTER ExternalAuthMiddleware.
func ActiveUserResolverMiddleware(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetActiveUser(r)

			// If no user or no external UID, skip resolution
			if user == nil || !user.HasExternalUID {
				next.ServeHTTP(w, r)
				return
			}

			// Attempt to resolve corporeal domain binding
			binding, err := ResolveCorporealDomainByExternalUID(r.Context(), db, user.ExternalUID)
			if err != nil {
				log.Printf("[auth] Failed to resolve corporeal domain: %v", err)
				// Continue without binding
				next.ServeHTTP(w, r)
				return
			}

			// Update user with binding info
			if binding.Found {
				user.Bound = true
				user.CorporealDomainID = binding.CorporealDomainID
				user.CorporealDomainUID = binding.CorporealDomainUID
			}

			// Continue with updated user
			next.ServeHTTP(w, r)
		})
	}
}
