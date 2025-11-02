package api

import (
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
	mux := s.mux

	// ============================================================
	//  CORE / SYSTEM ROUTES
	// ============================================================
	mux.HandleFunc("/api/ping", s.handlePing)
	mux.HandleFunc("/api/info", s.handleInfo)
	mux.HandleFunc("/api/health", s.handleHealth)

	// Status & domain metadata
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/domain/info", s.handleDomainInfo)

	// Identity management
	mux.HandleFunc("/api/identities", s.handleIdentities)
	mux.HandleFunc("/api/identity/list", s.handleIdentities)

	// ============================================================
	//  CANON / EXPORT (manual trigger)
	// ============================================================
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

	// ============================================================
	//  REGISTRY MODULES
	// ============================================================
	auth.Register(mux, s.db)
	identities.Register(mux, s.db)
	atlas.Register(mux, s.db)
	receipts.Register(mux, s.db)
	terra.Register(mux, s.db)

	// ============================================================
	//  BOOTSTRAP (Finagler ↔ DIS-Core integration)
	// ============================================================
	// All /api/bootstrap/* routes are defined inside internal/api/bootstrap.go
	s.RegisterBootstrapRoutes(s.Ledger.DB)

	// ============================================================
	//  ADDITIONAL ROUTE GROUPS
	// ============================================================
	registerFlowAPI(s)           // workflow orchestration
	s.registerImportListRoute()  // import list
	s.registerImportRoutes()     // import POST
	s.registerNetworkRoutes()    // peer/network layer
	s.registerDBRoutes()         // DB diagnostics
	s.registerVersionRoutes()    // version info
	s.registerMirrorSpinRoutes() // mirror spin test
	s.registerReconcileRoutes()  // reconciliation endpoints
}
