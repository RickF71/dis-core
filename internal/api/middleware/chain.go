package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"dis-core/internal/policy"
)

// Attach configures the universal middleware stack for all /api routes
// This function applies all middleware in the correct order:
// 1. Chi built-ins: Logger, Recoverer, RequestID
// 2. Custom middleware: DB, Provenance, PolicyEngine
func Attach(r *chi.Mux, pool *pgxpool.Pool, engine policy.PolicyEngine) {
	// Chi built-in middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Custom DIS-Core middleware
	r.Use(WithDB(pool))
	r.Use(WithProvenance())
	r.Use(WithPolicyEngine(engine))
}

// AttachWithCORS is an alternative that includes CORS middleware
func AttachWithCORS(r *chi.Mux, pool *pgxpool.Pool, engine policy.PolicyEngine) {
	// Chi built-in middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// CORS middleware (if needed)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Custom DIS-Core middleware
	r.Use(WithDB(pool))
	r.Use(WithProvenance())
	r.Use(WithPolicyEngine(engine))
}
