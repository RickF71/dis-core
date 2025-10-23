package api

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"dis-core/internal/config"
	"dis-core/internal/domain"
	"dis-core/internal/ledger"
	"dis-core/internal/overlay"
	"dis-core/internal/policy"
	"dis-core/internal/schema"
)

// WithLedger sets the Ledger pointer and returns the server (chainable)
func (s *Server) WithLedger(led *ledger.Ledger) *Server {
	s.Ledger = led
	return s
}

type Server struct {
	mux *http.ServeMux
	db  *sql.DB

	// Core components
	Ledger *ledger.Ledger

	// Managers (YAML import & domain logic)
	DomainManager  *domain.Manager
	SchemaManager  *schema.Manager
	PolicyManager  *policy.Manager
	OverlayManager *overlay.Manager // safe to keep even if stub

	// Optional legacy store field (some older routes expect it)
	Store *ledger.Store

	// Optional logger
	logger *log.Logger

	// Optional schema registry (for validation)
	schemas *schema.Registry
}

// Mux returns the internal HTTP mux for this server.
func (s *Server) Mux() *http.ServeMux { return s.mux }

// handlePing is a simple health endpoint for API status checks.
func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"message": "DIS node alive",
	})
}

// handleInfo reports basic build and version info.
func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version": "0.9.3",
		"core":    "DIS-Core",
	})
}

// handleHealth performs a simple self-check.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"health": "green",
	})
}

// NewServer constructs and initializes a DIS-Core API server.
func NewServer(cfg *config.Config, led *ledger.Ledger, db *sql.DB) *Server {
	s := &Server{
		mux: http.NewServeMux(),
		db:  db,
	}
	s.Store = ledger.NewStore(db)
	s.RegisterAllRoutes() // reconnect routes from routes.go

	// Do NOT wrap s.mux with CORS here; wrap at ListenAndServe

	return s
}

// WithLogger sets a custom logger and returns the server (chainable)
func (s *Server) WithLogger(l *log.Logger) *Server {
	s.logger = l
	return s
}

// WithSchemas sets a schema registry and returns the server (chainable)
func (s *Server) WithSchemas(reg *schema.Registry) *Server {
	s.schemas = reg
	return s
}

// handleSchemaList returns all registered schema IDs and versions as JSON.
func (s *Server) handleSchemaList(w http.ResponseWriter, r *http.Request) {
	if s.schemas == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "schema registry unavailable"})
		return
	}
	type schemaInfo struct {
		ID      string `json:"id"`
		Version string `json:"version"`
	}
	var out []schemaInfo
	for _, e := range s.schemasEntries() {
		out = append(out, schemaInfo{ID: e.ID, Version: e.Version})
	}
	writeJSON(w, http.StatusOK, out)
}

// schemasEntries returns all schema entries in the registry.
func (s *Server) schemasEntries() []schema.Entry {
	if s.schemas == nil {
		return nil
	}
	entries := make([]schema.Entry, 0, len(s.schemasEntriesMap()))
	for _, e := range s.schemasEntriesMap() {
		entries = append(entries, e)
	}
	return entries
}

// schemasEntriesMap returns the byKey map from the registry (read-only).
func (s *Server) schemasEntriesMap() map[string]schema.Entry {
	if s.schemas == nil {
		return nil
	}
	return s.schemas.ByKey()
}

func (s *Server) Run(ctx context.Context) error {
	// TODO: Start HTTP server, handle graceful shutdown
	return nil
}

// TODO: Keep RegisterAllRoutes() as canonical, and split per-route files as needed.
