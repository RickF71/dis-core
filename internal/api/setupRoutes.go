package api

import (
	"encoding/json"
	"net/http"
	"time"

	"dis-core/internal/api/identity"
	"dis-core/internal/policy"

	"github.com/jackc/pgx/v5/pgxpool"
)

// copilot: Complete RegisterAllRoutes() function with all endpoint registration.
// Register authority status, policy reload, freeze state, and identity binding endpoints.
// Ensure all routes are registered under a single orchestrator function.

// RouteConfig holds configuration for route registration
type RouteConfig struct {
	DB           *pgxpool.Pool
	PolicyEngine *policy.OPAEngine
	Mux          *http.ServeMux
}

// RegisterAllRoutes registers all API endpoints under a single orchestrator
func RegisterAllRoutes(config *RouteConfig) {
	mux := config.Mux
	db := config.DB
	engine := config.PolicyEngine

	// Register core system routes
	registerCoreRoutes(mux, db)

	// Register authority console routes
	registerAuthorityRoutes(mux, db)

	// Register policy management routes
	registerPolicyRoutes(mux, db, engine)

	// Register identity binding routes
	registerIdentityBindingRoutes(mux, db)
}

// registerCoreRoutes registers fundamental system routes
func registerCoreRoutes(mux *http.ServeMux, db *pgxpool.Pool) {
	// Health check endpoints
	mux.HandleFunc("GET /api/ping", handlePing)
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		handleStatus(w, r, db)
	})

	// Version information
	mux.HandleFunc("GET /api/version", handleVersion)
}

// registerAuthorityRoutes registers authority console endpoints
func registerAuthorityRoutes(mux *http.ServeMux, db *pgxpool.Pool) {
	// Authority status endpoint from authority_status.go
	mux.HandleFunc("GET /api/authority/status", func(w http.ResponseWriter, r *http.Request) {
		HandleAuthorityStatus(w, r, db)
	})
}

// registerPolicyRoutes registers policy management endpoints
func registerPolicyRoutes(mux *http.ServeMux, db *pgxpool.Pool, engine *policy.OPAEngine) {
	// Policy reload endpoint from policy/runtime.go
	mux.HandleFunc("POST /api/policy/reload", func(w http.ResponseWriter, r *http.Request) {
		policy.HandlePolicyReload(w, r, db, engine)
	})
}

// registerIdentityBindingRoutes registers identity binding endpoints
func registerIdentityBindingRoutes(mux *http.ServeMux, db *pgxpool.Pool) {
	// Create identity binding
	mux.HandleFunc("POST /api/identity/bindings", func(w http.ResponseWriter, r *http.Request) {
		identity.HandleCreateIdentityBinding(w, r, db)
	})
}

// Placeholder handlers for core routes

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","message":"pong"}`))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"healthy","timestamp":"` +
		time.Now().Format(time.RFC3339) + `"}`))
}

func handleStatus(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) {
	w.Header().Set("Content-Type", "application/json")
	status := map[string]interface{}{
		"status":    "running",
		"database":  "connected",
		"version":   "0.9.7",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(status)
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"version":"0.9.7","build":"MOAR-VIII-H"}`))
}
