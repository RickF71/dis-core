package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"dis-core/internal/core/session"
)

// SessionAuthMiddleware checks Authorization: Bearer <token> header and, if present,
// validates the session and attaches an ActiveUser bound to the corporeal domain.
func SessionAuthMiddleware(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := strings.TrimSpace(r.Header.Get("Authorization"))
			if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
				// no bearer token, continue
				next.ServeHTTP(w, r)
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Validate session inside a tx
			ctx := r.Context()
			pool := db
			tx, err := pool.Begin(ctx)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			defer func() { _ = tx.Rollback(ctx) }()

			sess, err := session.LoadSessionTx(ctx, tx, token)
			if err != nil {
				// invalid token -> continue without binding (handler will return 401)
				next.ServeHTTP(w, r)
				return
			}

			// reject revoked sessions explicitly
			if sess.RevokedAt != nil {
				// Session has been revoked - return 401 Unauthorized
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				http.Error(w, "session revoked", http.StatusUnauthorized)
				return
			}
			// reject expired
			if time.Now().UTC().After(sess.ExpiresAt) {
				next.ServeHTTP(w, r)
				return
			}

			// Commit tx — we only read but keep tx semantics consistent
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
			// Also attach active actor seat id (from session)
			ctx3 := context.WithValue(ctx2, activeActorKey, sess.SeatID)
			next.ServeHTTP(w, r.WithContext(ctx3))
		})
	}
}
