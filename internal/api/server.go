package api

import (
	"encoding/json"
	"log"
	"net/http"

	"dis-core/internal/authority"
	"dis-core/internal/ledger"
	"dis-core/internal/policy"
	"dis-core/internal/schema"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Server is the core API instance for DIS-Core.
type Server struct {
	db      *pgxpool.Pool
	ledger  *ledger.Ledger
	mux     *http.ServeMux
	schemas *schema.Registry
	logger  *log.Logger
	policy  policy.PolicyEngine

	authoritySchema  *authority.AuthoritySchema
	authorityConsole *authority.Console
}

// New creates a new Server instance and wires all dependencies.
func New(db *pgxpool.Pool, led *ledger.Ledger) *Server {
	s := &Server{
		db:     db,
		ledger: led,
		mux:    http.NewServeMux(),
		logger: log.Default(),
	}

	authzSchema, err := authority.LoadSchema("./internal/schema/schema.json")
	if err != nil {
		s.logger.Printf("failed to load authority schema: %v", err)
	}

	s.authoritySchema = authzSchema
	s.authorityConsole = authority.NewConsole(db, led, nil, s.authoritySchema)
	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/api/ping", s.handlePing)
	s.mux.HandleFunc("/api/status", s.handleStatus)
	s.mux.HandleFunc("/api/domain/", s.handleDomainGet)

	// Register admin routes
	s.registerAdminRoutes()
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "running",
		"service":   "dis-core",
		"ledger":    s.ledger != nil,
		"authority": s.authorityConsole != nil,
		"schemas":   s.schemas != nil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleDomainGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	domainID := r.URL.Path[len("/api/domain/"):]
	if domainID == "" {
		http.Error(w, "domain ID required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	var domainData string
	err := s.db.QueryRow(ctx, `
SELECT content FROM canon
WHERE type = 'domain' AND content->>'domain_id' = $1
LIMIT 1
`, domainID).Scan(&domainData)

	if err != nil {
		s.logger.Printf("domain query failed: %v", err)
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(domainData))
}

func (s *Server) Handler() *http.ServeMux { return s.mux }
func (s *Server) DB() *pgxpool.Pool       { return s.db }
func (s *Server) Ledger() *ledger.Ledger  { return s.ledger }
func (s *Server) Mux() *http.ServeMux     { return s.mux }
