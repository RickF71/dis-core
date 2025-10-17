package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"dis-core/internal/api/atlas"
	"dis-core/internal/config"
	"dis-core/internal/db"
	"dis-core/internal/policy"
)

// Server represents the DIS-CORE REST node.
type Server struct {
	store    *sql.DB
	cfg      *config.Config
	policy   *policy.Policy
	sum      string
	coreHash string
	mux      *http.ServeMux
	atlas    *atlas.AtlasStore
}

func (s *Server) AttachAtlas(a *atlas.AtlasStore) {
	s.atlas = a
}

// NewServer constructs a new REST server instance.
func NewServer(store *sql.DB, cfg *config.Config, pol *policy.Policy, sum string, coreHash string) *Server {
	s := &Server{
		store:    store,
		cfg:      cfg,
		policy:   pol,
		sum:      sum,
		coreHash: coreHash,
		mux:      http.NewServeMux(),
	}

	atlasStore, err := atlas.InitAtlasStore(store)
	if err != nil {
		log.Fatalf("failed to init atlas store: %v", err)
	}
	s.atlas = atlasStore

	// register modular routes
	s.RegisterAllRoutes()

	return s
}

// Start launches the REST API server for DIS-CORE.
func (s *Server) Start(addr string) error {
	log.Printf("🛰️  DIS-CORE REST API listening on %s\n", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      WithCORS(s.mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server.ListenAndServe()
}

func (s *Server) Mux() *http.ServeMux {
	return s.mux
}

// --- Core handlers ---
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   db.NowRFC3339Nano(),
	})
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"core_hash": s.coreHash,
		"policy":    s.sum,
		"db_path":   s.cfg.DatabaseDSN,
	})
}

// --- JSON utility ---
func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("❌ JSON encode error: %v", err)
	}
}
