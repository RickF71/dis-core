package api

import (
	"database/sql"
	"log"
	"net/http"

	"dis-core/internal/canon"
	"dis-core/internal/registry/atlas"
	"dis-core/internal/registry/auth"
	"dis-core/internal/registry/identities"
	"dis-core/internal/registry/receipts"
	"dis-core/internal/registry/terra"
)

// RegisterAPIs wires all endpoint groups into the server mux.
func (s *Server) RegisterAPIs() {
	// Register /api/eval if PolicyEngine is set
	// Register /api/eval if PolicyEngine is set (interface nil check)
	// Register /api/eval if PolicyEngine is set (interface nil check)
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

		if err := canon.ExportDomains(s.db, "domains/_auto"); err != nil {
			log.Printf("⚠️ Canon export failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","message":"Canonical export complete"}`))
		log.Println("✅ Canon export triggered manually via /api/canon/export")
	})

	// Modular packages
	auth.Register(mux, s.db)
	identities.Register(mux, s.db)
	atlas.Register(mux, s.db)
	receipts.Register(mux, s.db)
	terra.Register(mux, s.db)

	// Register import receipts list route
	s.registerImportListRoute()

	// Register import (POST) route
	s.registerImportRoutes()

	// Register network API routes
	s.registerNetworkRoutes()
	//log.Printf("✅ Registered route: /api/net/peers")

	s.registerDBRoutes() //

	s.registerVersionRoutes()
	s.registerMirrorSpinRoutes() //
	//s.registerStatusRoutes()

}

// Register network API routes
func BuildMux(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	RegisterStatusRoutes(mux)
	//RegisterCanonRoutes(mux, db)
	RegisterDomainRoutes(mux, db)
	return mux
}
