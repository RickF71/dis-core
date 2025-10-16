package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"dis-core/internal/config"
	"dis-core/internal/db"
	"dis-core/internal/policy"
	// ✅ Added
)

// Server represents the DIS-CORE REST node.
type Server struct {
	store    *sql.DB
	cfg      *config.Config
	policy   *policy.Policy
	sum      string
	coreHash string
	mux      *http.ServeMux
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

	// Register all routes
	s.registerRoutes()
	return s
}

// Start launches the REST API server for DIS-CORE.
func (s *Server) Start(addr string) error {
	log.Printf("🛰️  DIS-CORE REST API listening on %s\n", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      WithCORS(s.mux), // ✅ global CORS middleware applied here
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server.ListenAndServe()
}

// --- Route registration ---
func (s *Server) registerRoutes() {
	// === Health ===
	s.mux.HandleFunc("/healthz", s.handleHealth)

	// === Core info ===
	s.mux.HandleFunc("/ping", s.handlePing)
	s.mux.HandleFunc("/info", s.handleInfo)
	s.mux.HandleFunc("/verify", HandleExternalVerify)
	s.mux.HandleFunc("/receipts", s.handleReceipts)

	// === Auth / Identity ===
	s.mux.HandleFunc("/api/auth/revoke", s.HandleAuthRevoke)
	s.mux.HandleFunc("/api/auth/handshake", s.HandleDISAuthHandshake)
	s.mux.HandleFunc("/api/auth/virtual_usgov", HandleVirtualUSGovCredential)
	s.mux.HandleFunc("/api/identity/register", HandleIdentityRegister(s.store))
	s.mux.HandleFunc("/api/identity/list", HandleIdentityList(s.store))
	s.mux.HandleFunc("/api/overlay/", GetOverlayHandler) // mounts /api/overlay/:domain/:scope

	// === v0.9.3 self-maintenance ===
	RegisterConsoleAuthRoutes(s.mux)
	RegisterStatusRoutes(s.mux)

	// === Terra sync API ===
	RegisterTerraRoutes(s.mux)

	// === Root ===
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "🌐 DIS-CORE v0.9.3 — Self-Maintenance and Reflexive Identity\nTime: %s\n", db.NowRFC3339Nano())
	})
}

// --- Core Handlers ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Keep it simple and fast—used by load balancers & UIs
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
		"db_path":   s.cfg.DatabasePath,
	})
}

func (s *Server) handleReceipts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	list, err := db.ListReceipts(s.store, db.ListOpts{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// --- Utility JSON writer ---
func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("❌ JSON encode error: %v", err)
	}
}
