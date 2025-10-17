package api

import (
	"net/http"

	"dis-core/internal/api/atlas"
	"dis-core/internal/api/auth"
	"dis-core/internal/api/identities"
	"dis-core/internal/api/receipts"
	"dis-core/internal/api/terra"
)

// RegisterAllRoutes wires all endpoint groups into the server mux.
func (s *Server) RegisterAllRoutes() {
	mux := s.mux

	// Core system routes
	mux.HandleFunc("/verify", HandleVerify) // or s.handleVerify if you make it a method
	mux.HandleFunc("/ping", s.handlePing)
	mux.HandleFunc("/info", s.handleInfo)
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/identities", s.handleIdentities)

	// Modular packages
	auth.Register(mux, s.store)
	identities.Register(mux, s.store)
	atlas.Register(mux, s.store)
	receipts.Register(mux, s.store)
	terra.Register(mux, s.store)

	// Root fallback
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("🌐 DIS-CORE v0.9.3 — Self-Maintenance and Reflexive Identity\n"))
	})
}
