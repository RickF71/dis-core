package auth

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"dis-core/internal/core/session"
)

// SessionAuthMiddleware reads the `dis_session` cookie (if present), validates
// the session against the DB and, when valid, attaches an ActiveUser bound to
// the corporeal domain. Missing, stale or invalid cookies are treated as
// anonymous (the request continues without an authenticated user).
func SessionAuthMiddleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First, prefer Authorization: Bearer <token> if present (preserves
			// existing API token behavior). If not present, fall back to the
			// cookie-based session flow.
			authz := r.Header.Get("Authorization")
			if authz != "" {
				// Preserve original bearer-token behavior
				authz = strings.TrimSpace(authz)
				if strings.HasPrefix(authz, "Bearer ") {
					token := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
					if token == "" {
						next.ServeHTTP(w, r)
						return
					}

					log.Printf("[session-mw] Authorization bearertoken present (len=%d)", len(token))

					ctx := r.Context()
					tx, err := pool.Begin(ctx)
					if err != nil {
						next.ServeHTTP(w, r)
						return
					}
					defer func() { _ = tx.Rollback(ctx) }()

					sess, err := session.LoadSessionTx(ctx, tx, token)
					if err != nil {
						next.ServeHTTP(w, r)
						return
					}
					if sess.RevokedAt != nil {
						w.Header().Set("Content-Type", "text/plain; charset=utf-8")
						http.Error(w, "session revoked", http.StatusUnauthorized)
						return
					}
					if time.Now().UTC().After(sess.ExpiresAt) {
						next.ServeHTTP(w, r)
						return
					}
					_ = tx.Commit(ctx)

					user := &ActiveUser{
						ExternalUID:        "",
						HasExternalUID:     false,
						Bound:              true,
						CorporealDomainUID: sess.DomainID,
						CorporealDomainID:  0,
					}
					ctx2 := WithActiveUser(r.Context(), user)
					ctx3 := context.WithValue(ctx2, activeActorKey, sess.SeatID)
					next.ServeHTTP(w, r.WithContext(ctx3))
					return
				}
			}

			// No Authorization header — try cookie-based session
			c, err := r.Cookie("dis_session")
			if err != nil {
				// No cookie → anonymous
				next.ServeHTTP(w, r)
				return
			}
			token := c.Value
			log.Printf("[session-mw] cookie dis_session present (len=%d)", len(token))
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Load session from DB inside a transaction
			ctx := r.Context()
			tx, err := pool.Begin(ctx)
			if err != nil {
				// DB error — treat as anonymous
				next.ServeHTTP(w, r)
				return
			}
			defer func() { _ = tx.Rollback(ctx) }()

			sess, err := session.LoadSessionTx(ctx, tx, token)
			if err != nil || sess == nil {
				// Invalid or not found -> anonymous
				next.ServeHTTP(w, r)
				return
			}

			// Reject revoked or expired sessions
			if sess.RevokedAt != nil {
				// revoked -> treat as anonymous
				next.ServeHTTP(w, r)
				return
			}
			if time.Now().UTC().After(sess.ExpiresAt) {
				// expired -> anonymous
				next.ServeHTTP(w, r)
				return
			}

			// Commit tx (we only read but keep semantics)
			_ = tx.Commit(ctx)

			// Attach a minimal ActiveUser bound to the corporeal domain
			user := &ActiveUser{
				ExternalUID:        "",
				HasExternalUID:     false,
				Bound:              true,
				CorporealDomainUID: sess.DomainID,
				CorporealDomainID:  0,
			}
			ctx2 := WithActiveUser(r.Context(), user)

			// Attach active actor seat id if present
			ctx3 := ctx2
			if sess.SeatID != "" {
				ctx3 = context.WithValue(ctx3, activeActorKey, sess.SeatID)
			}

			next.ServeHTTP(w, r.WithContext(ctx3))
		})
	}
}
