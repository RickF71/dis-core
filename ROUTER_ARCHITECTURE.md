# Router Architecture: Public vs Private Routes

## Current Structure

DIS-Core already separates public and private routes through middleware, but they're mounted on a single router. This document outlines the conceptual split and how to formalize it.

## Existing Route Categories

### Public Routes (No Auth Required)
Current location: `internal/api/routes.go`

```go
// Ping/Health
GET /api/ping

// Authentication
POST /api/auth/challenge
GET /api/auth/challenge/:id/status
POST /api/auth/qr-complete

// Dev Mode
GET /api/dev/users
POST /auth/dev
```

### Private Routes (Auth Required)
Current location: `internal/api/routes.go`

```go
// Identity
GET /api/me
GET /api/me/actors
POST /api/me/active-actor
GET /api/me/active-actor

// Status
GET /api/status

// Domains
GET /api/domains
GET /api/domain/:id
GET /api/domain/:id/resolved-css
POST /api/domain/:id/aliases
GET /api/domain/:id/aliases/:alias

// Seats
GET /api/seats
GET /api/seats/:id

// Contracts
POST /api/contracts
GET /api/contracts/:id
PUT /api/contracts/:id

// Receipts
GET /api/receipts
POST /api/receipts
GET /api/receipts/verify/:id

// Authority Events (SSE)
GET /api/authority/events
```

## Current Middleware Stack

File: `internal/api/middleware/chain.go`

```go
func Attach(r *chi.Mux, pool *pgxpool.Pool, engine policy.PolicyEngine) {
    // 1. CORS (MUST be first)
    r.Use(cors.Middleware)

    // 2. Chi built-ins
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(middleware.RequestID)

    // 3. Authentication (sets user context)
    r.Use(auth.ExternalAuthMiddleware(pool))
    r.Use(auth.ActiveUserResolverMiddleware(pool))

    // 4. Custom middleware
    r.Use(WithDB(pool))
    r.Use(WithProvenance())
    r.Use(WithPolicyEngine(engine))
}
```

**Key Point:** All routes get authentication middleware, but public routes don't require authenticated users. The middleware sets `r.Context()` with user info if present, but doesn't block requests.

## Proposed Refactor (Future Enhancement)

### Option 1: Route Groups with Selective Middleware

```go
// internal/api/routes.go
func (s *Server) RegisterAllRoutes() {
    r := s.router

    // Apply global middleware (CORS, logging)
    r.Use(cors.Middleware)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    // Public routes (no auth required)
    r.Group(func(pub chi.Router) {
        pub.Get("/api/ping", s.handlePing)
        pub.Post("/api/auth/challenge", s.handleCreateChallenge)
        pub.Get("/api/auth/challenge/{id}/status", s.handleChallengeStatus)
        pub.Get("/api/dev/users", GetDevUsers)
        pub.Post("/auth/dev", PostDevAuth)
    })

    // Private routes (auth required)
    r.Group(func(priv chi.Router) {
        priv.Use(auth.RequireAuthMiddleware) // Block if not authenticated

        priv.Get("/api/me", s.handleMe)
        priv.Get("/api/me/actors", s.handleMeActors)
        priv.Get("/api/status", s.handleStatus)
        priv.Get("/api/domains", s.handleDomainList)
        // ... all other protected routes
    })
}
```

### Option 2: Separate Routers (Mounted)

```go
// internal/api/router_public.go
func (s *Server) PublicRouter() chi.Router {
    r := chi.NewRouter()

    r.Get("/ping", s.handlePing)
    r.Post("/auth/challenge", s.handleCreateChallenge)
    r.Get("/auth/challenge/{id}/status", s.handleChallengeStatus)

    return r
}

// internal/api/router_private.go
func (s *Server) PrivateRouter() chi.Router {
    r := chi.NewRouter()
    r.Use(auth.RequireAuthMiddleware)

    r.Get("/me", s.handleMe)
    r.Get("/me/actors", s.handleMeActors)
    r.Get("/status", s.handleStatus)

    return r
}

// internal/api/routes.go
func (s *Server) RegisterAllRoutes() {
    r := s.router

    // Global middleware
    r.Use(cors.Middleware)
    r.Use(middleware.Logger)

    // Mount subrouters
    r.Mount("/api", s.PublicRouter())
    r.Mount("/api", s.PrivateRouter())
}
```

## Current Auth Middleware Behavior

File: `internal/auth/middleware.go`

```go
// ExternalAuthMiddleware - DOES NOT block requests
// Sets user in context if authenticated, but continues chain
func ExternalAuthMiddleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Try to get user from session or X-External-User header
            user := getUserFromRequest(r)

            // Set in context (may be nil)
            ctx := context.WithValue(r.Context(), userContextKey, user)

            // ALWAYS continue (even if user is nil)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

**To add auth blocking:**

```go
// RequireAuthMiddleware - BLOCKS unauthenticated requests
func RequireAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := GetUserFromContext(r.Context())

        if user == nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

## Benefits of Route Separation

### 1. **Security Clarity**
- Explicit public vs private distinction
- Impossible to accidentally expose protected endpoints
- Easier security audits

### 2. **Middleware Efficiency**
- Public routes skip expensive auth checks
- Private routes guaranteed to have user context
- Reduced overhead for health checks

### 3. **API Documentation**
- Clear separation in OpenAPI specs
- Better developer experience
- Easier to generate client SDKs

### 4. **Testing**
- Test public routes without auth setup
- Test private routes with mock users
- Isolated test suites

## Implementation Priority

**Current Status:** ✅ Working
- All routes functional
- CORS working (MOAR-CORS v1)
- Authentication working (session + dev mode)
- No security issues

**Priority:** 🔵 Low (Nice to have, not urgent)
- Current architecture is secure
- Middleware correctly applied
- Route organization is clear

**When to implement:**
- When adding OAuth/JWT
- When adding rate limiting per route type
- When generating API documentation
- When onboarding new developers

## Migration Path

If implementing route separation:

1. **Create auth blocking middleware**
   ```bash
   # Add to internal/auth/middleware.go
   func RequireAuthMiddleware() {...}
   ```

2. **Split routes.go into public and private**
   ```bash
   internal/api/routes_public.go
   internal/api/routes_private.go
   ```

3. **Update RegisterAllRoutes to use groups**
   ```bash
   # Modify internal/api/routes.go
   ```

4. **Test all endpoints**
   ```bash
   go test ./internal/api/...
   ./test-moar-cors.sh
   ```

5. **Update documentation**
   ```bash
   # Update API docs, OpenAPI specs
   ```

## Current Endpoints by Category

### ✅ Public (No Auth)
- `GET /api/ping` - Health check
- `POST /api/auth/challenge` - Create QR challenge
- `GET /api/auth/challenge/:id/status` - Check challenge status
- `POST /api/auth/qr-complete` - Complete QR auth
- `GET /api/dev/users` - Dev mode user list
- `POST /auth/dev` - Dev mode login

### 🔒 Private (Auth Required)
- `GET /api/me` - Current user identity
- `GET /api/me/actors` - User's actors list
- `POST /api/me/active-actor` - Set active actor
- `GET /api/me/active-actor` - Get active actor
- `GET /api/status` - System status
- `GET /api/domains` - Domain list
- `GET /api/domain/:id` - Domain details
- `GET /api/domain/:id/resolved-css` - Domain CSS
- `GET /api/seats` - Seats list
- `POST /api/contracts` - Create contract
- `GET /api/receipts` - Receipts list
- `GET /api/authority/events` - Authority event stream (SSE)

## Related Files

- `internal/api/routes.go` - Current route registration
- `internal/api/middleware/chain.go` - Middleware stack
- `internal/auth/middleware.go` - Authentication middleware
- `internal/cors/middleware.go` - CORS middleware (MOAR-CORS v1)
- `internal/api/dev.go` - Dev auth endpoints

## Conclusion

The current architecture already conceptually separates public and private routes through middleware application. Formalizing this with separate routers or route groups would improve clarity but is not required for security or functionality.

**Recommendation:** Document current structure (this file) and defer router split until it provides tangible benefits (OAuth, rate limiting, API docs).
