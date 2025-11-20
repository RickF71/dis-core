package middleware

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"dis-core/internal/auth"
	"dis-core/internal/cors"
	"dis-core/internal/policy"
)

// Attach configures the universal middleware stack for all /api routes
// This function applies all middleware in the correct order:
// 1. CORS (MUST be first for preflight requests) - using MOAR-CORS v1
// 2. Chi built-ins: Logger, Recoverer, RequestID
// 3. External authentication: ExternalAuth, ActiveUserResolver
// 4. Custom middleware: DB, Provenance, PolicyEngine
func Attach(r *chi.Mux, pool *pgxpool.Pool, engine policy.PolicyEngine) {
	// CORS middleware (MUST be first to handle OPTIONS preflight)
	// Using MOAR-CORS v1 with dynamic origin checking and DIS_ALLOWED_ORIGINS support
	r.Use(cors.Middleware)

	// Chi built-in middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// External authentication stack (MUST be early in chain)
	r.Use(auth.ExternalAuthMiddleware(pool)) // Now supports session-based auth
	// Session token middleware: checks Authorization: Bearer <token> and attaches ActiveUser
	r.Use(auth.SessionAuthMiddleware(pool))
	r.Use(auth.ActiveUserResolverMiddleware(pool))

	// Custom DIS-Core middleware
	r.Use(WithDB(pool))
	r.Use(WithProvenance())
	r.Use(WithPolicyEngine(engine))
}

// AttachWithCORS is deprecated - CORS is now included in Attach()
// Kept for backward compatibility, but just calls Attach()
func AttachWithCORS(r *chi.Mux, pool *pgxpool.Pool, engine policy.PolicyEngine) {
	Attach(r, pool, engine)
}
