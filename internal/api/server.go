package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"dis-core/internal/config"
	"dis-core/internal/db"
	"dis-core/internal/policy"
)

// Server represents the DIS-PERSONAL REST node.
// For v0.8, it uses a direct *sql.DB connection.
// In v0.9+, this can be replaced by a higher-level db.Store wrapper.
type Server struct {
	store    *sql.DB
	cfg      *config.Config
	policy   *policy.Policy
	sum      string
	coreHash string
}

// NewServer constructs a new REST server instance.
func NewServer(store *sql.DB, cfg *config.Config, pol *policy.Policy, sum string, coreHash string) *Server {
	return &Server{
		store:    store,
		cfg:      cfg,
		policy:   pol,
		sum:      sum,
		coreHash: coreHash,
	}
}

// Start launches the REST API server for DIS-PERSONAL.
func (s *Server) Start(addr string) error {
	http.HandleFunc("/ping", s.handlePing)
	http.HandleFunc("/info", s.handleInfo)
	http.HandleFunc("/verify", HandleExternalVerify) // now implemented as a stub below
	http.HandleFunc("/receipts", s.handleReceipts)
	http.HandleFunc("/api/auth/revoke", s.HandleAuthRevoke)

	// --- NEW: DIS-Auth + Virtual USGOV endpoints ---
	http.HandleFunc("/api/auth/handshake", s.HandleDISAuthHandshake)
	http.HandleFunc("/api/auth/virtual_usgov", HandleVirtualUSGovCredential)
	http.HandleFunc("/api/auth/console", HandleConsoleAuth)
	// (optional later) http.HandleFunc("/api/auth/revoke", HandleAuthRevoke)
	http.HandleFunc("/api/identity/register", HandleIdentityRegister(s.store))
	http.HandleFunc("/api/identity/list", HandleIdentityList(s.store))

	log.Printf("🛰️  DIS-PERSONAL REST API listening on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// --- Handlers ---

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
	// Load query params
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	// Fetch from DB
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
