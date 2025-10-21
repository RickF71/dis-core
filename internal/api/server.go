package api

import (
	"database/sql"
	"dis-core/internal/api/atlas"
	"dis-core/internal/schema"
	"log"
	"net/http"

	disnet "dis-core/internal/net"
)

// Server represents the running DIS-Core API node.
type Server struct {
	store      *sql.DB
	logger     *log.Logger
	schemas    *schema.Registry
	mux        *http.ServeMux
	atlas      *atlas.AtlasStore
	NetManager *disnet.Manager // 👈 restored network manager
	Version    string
	CoreHash   string
}

// Mux returns the internal HTTP mux for this server.
func (s *Server) Mux() *http.ServeMux {
	return s.mux
}

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
func NewServer(store *sql.DB) *Server {
	s := &Server{
		store:      store,
		mux:        http.NewServeMux(),
		NetManager: disnet.NewManager(store), // 👈 initialize manager
		logger:     log.Default(),
	}

	s.RegisterAllRoutes()
	s.registerMirrorSpinRoutes()
	s.mux.HandleFunc("/api/schema/list", s.handleSchemaList)
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
