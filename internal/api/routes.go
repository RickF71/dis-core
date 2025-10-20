package api

import (
	"database/sql"
	"log"
	"net/http"

	"dis-core/internal/api/atlas"
	"dis-core/internal/api/auth"
	"dis-core/internal/api/identities"
	"dis-core/internal/api/receipts"
	"dis-core/internal/api/terra"
	"dis-core/internal/canon"
)

// RegisterAllRoutes wires all endpoint groups into the server mux.
func (s *Server) RegisterAllRoutes() {
	mux := s.mux

	// Core system routes
	//mux.HandleFunc("/verify", s.handleVerify)
	mux.HandleFunc("/ping", s.handlePing)
	// Backwards-compatible API routes used by Finagler frontend
	mux.HandleFunc("/api/status", s.handlePing)
	mux.HandleFunc("/api/domain/info", s.handleDomainInfo)
	mux.HandleFunc("/info", s.handleInfo)
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/identities", s.handleIdentities)
	mux.HandleFunc("/api/identity/list", s.handleIdentities)
	// Canon export (manual trigger)
	mux.HandleFunc("/api/canon/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := canon.ExportDomains(s.store, "domains/_auto"); err != nil {
			log.Printf("⚠️ Canon export failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","message":"Canonical export complete"}`))
		log.Println("✅ Canon export triggered manually via /api/canon/export")
	})

	// Modular packages
	auth.Register(mux, s.store)
	identities.Register(mux, s.store)
	atlas.Register(mux, s.store)
	receipts.Register(mux, s.store)
	terra.Register(mux, s.store)

	// Register network API routes
	s.registerNetworkRoutes()
	log.Printf("✅ Registered route: /api/net/peers")

	s.registerVersionRoutes()
	//s.registerStatusRoutes()

	// Root fallback
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("🌐 DIS-CORE v0.9.3 — Self-Maintenance and Reflexive Identity\n"))
	})
}

// Register network API routes
func BuildMux(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterStatusRoutes(mux)
	//RegisterCanonRoutes(mux, db)
	RegisterDomainRoutes(mux, db)
	return mux
}
