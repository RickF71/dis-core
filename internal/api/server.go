package api

import (
	"encoding/json"
	"log"
	"net/http"

	"dis-core/internal/ledger"
	"dis-core/internal/policy"
	"dis-core/internal/schema"

	"github.com/jackc/pgx/v5"
)

// TODO: Refactor Server struct to idiomatic Go style.
// - Use lowercase field names for internal-only data (db, ledger, mux, etc).
// - Keep fields unexported since this package is internal.
// - Provide exported getter methods only where needed (e.g., Handler()).
// - Ensure all references in other files match the new lowercase fields.
// Server is the core API instance for DIS-Core.
type Server struct {
	db      *pgx.Conn
	ledger  *ledger.Ledger
	mux     *http.ServeMux
	schemas *schema.Registry
	logger  *log.Logger

	policy policy.PolicyEngine
}

func New(db *pgx.Conn, led *ledger.Ledger) *Server {
	s := &Server{
		db:     db,
		ledger: led,
		mux:    http.NewServeMux(),
		logger: log.Default(),
	}

	// --- Policy engine wiring ---
	modules := map[string]string{
		"gates.rego":  "package dis.gates\n default allow = true", // stub content
		"risk.rego":   "package dis.risk\n default risk = 0",
		"freeze.rego": "package dis.freeze\n default frozen = false",
	}
	eng, err := policy.NewEngine(modules)
	if err != nil {
		s.logger.Printf("policy engine init failed: %v", err)
	} else {
		s.policy = eng
		s.logger.Println("policy engine initialized")
	}

	s.registerRoutes()
	return s
}

// registerRoutes sets up all the API routes
func (s *Server) registerRoutes() {
	s.RegisterAPIs()
}

// RegisterRoutes is an alias for registerRoutes for backward compatibility
func (s *Server) RegisterRoutes() {
	s.registerRoutes()
}

// HandleBootstrapStatus provides a bootstrap status endpoint
func (s *Server) HandleBootstrapStatus(w http.ResponseWriter, r *http.Request) {
	// Basic status response
	resp := map[string]interface{}{
		"status":       "ok",
		"bootstrapped": true,
		"timestamp":    "2024-11-07T00:00:00Z",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handlePolicyTest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	input := map[string]interface{}{
		"action": "domain.freeze.v1",
		"user":   "rick",
	}
	decision, err := s.policy.EvaluateAction(ctx, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(decision)
}

// Handler returns the underlying HTTP mux for ListenAndServe.
func (s *Server) Handler() *http.ServeMux { return s.mux }

// DB returns the database connection
func (s *Server) DB() *pgx.Conn { return s.db }

// Ledger returns the ledger instance
func (s *Server) Ledger() *ledger.Ledger { return s.ledger }

// Mux returns the HTTP mux (alias for Handler for backward compatibility)
func (s *Server) Mux() *http.ServeMux { return s.mux }
