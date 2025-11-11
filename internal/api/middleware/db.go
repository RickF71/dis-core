package middleware

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// contextKey is a custom type to avoid context key collisions
type contextKey string

const dbContextKey contextKey = "db"

// WithDB middleware injects a database connection pool into the request context
func WithDB(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Inject the database pool into the request context
			ctx := context.WithValue(r.Context(), dbContextKey, pool)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromDB retrieves the database connection pool from the request context
func FromDB(ctx context.Context) *pgxpool.Pool {
	if pool, ok := ctx.Value(dbContextKey).(*pgxpool.Pool); ok {
		return pool
	}
	return nil
}
