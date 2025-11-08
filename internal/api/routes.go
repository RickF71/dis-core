package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"dis-core/internal/api/auth"
	"dis-core/internal/api/identities"
	"dis-core/internal/registry/atlas"
	"dis-core/internal/registry/receipts"
	"dis-core/internal/registry/terra"
)

// RegisterAPIs wires all endpoint groups into the server mux.
func (s *Server) RegisterAPIs() {
	mux := s.mux

	// ============================================================
	// CORE / SYSTEM ROUTES
	// ============================================================
	mux.HandleFunc("GET /api/ping", s.handlePing)
	mux.HandleFunc("GET /api/status", s.HandleStatus)
	mux.HandleFunc("GET /api/bootstrap/status", s.HandleBootstrapStatus)

	// ============================================================
	// DOMAIN ROUTES (JSONB MODEL)
	// ============================================================

	// Create new domain (strict UUID model)
	mux.HandleFunc("POST /api/domain", s.handleCreateDomain)

	// Single domain fetch (returns full JSONB domain object)
	mux.HandleFunc("GET /api/domain/{id}", s.handleGetDomain)

	// Domain collection (list all domains)
	mux.HandleFunc("GET /api/domains", s.handleDomainList)

	// List all domains
	mux.HandleFunc("GET /api/domain", s.handleListDomains)
	mux.HandleFunc("PUT /api/domain/{id}", s.handleUpdateDomain)
	mux.HandleFunc("GET /api/domain/default", s.handleGetDefaultDomain)
	// Domain Files API
	// Domain-scoped Files API (separate from bootstrap /api/files)
	mux.HandleFunc("GET /api/domain/{id}/files", s.handleDomainFilesList)
	mux.HandleFunc("GET /api/domain/{id}/file/{filename}", s.handleDomainFileGet)
	mux.HandleFunc("PUT /api/domain/{id}/file/{filename}", s.handleDomainFilePut)
	mux.HandleFunc("DELETE /api/domain/{id}/file/{filename}", s.handleDomainFileDelete)
	mux.HandleFunc("POST /api/domain/{id}/file/{filename}", s.handleDomainFileCreate)
	mux.HandleFunc("GET /api/domain/{id}/announce", s.handleDomainAnnounce)
	mux.HandleFunc("GET /api/domain/{id}/css", s.handleDomainCSS)
	mux.HandleFunc("POST /api/domain/{id}/css", s.handleUpdateDomainCSS)
	mux.HandleFunc("POST /api/domain/{id}/file/rename", s.handleDomainFileRename)

	//mux.HandleFunc("POST /api/domain/{id}/file/{filename}/archive", s.handleArchiveDomainFile)
	//mux.HandleFunc("PUT /api/domain/{id}/file/{filename}", s.handleUpdateDomainFile)

	// jikka routes
	mux.HandleFunc("GET /api/jikka/{id}", s.handleGetJikka)
	mux.HandleFunc("POST /api/jikka", s.handleCreateJikka)
	mux.HandleFunc("GET /api/jikka/list", s.handleListJikkas)

	// ============================================================
	// IDENTITY & REGISTRY MODULES
	// ============================================================
	auth.Register(mux, s.DB())
	identities.Register(mux, s.DB())
	atlas.Register(mux, s.DB())
	receipts.Register(mux, s.DB())
	terra.Register(mux, s.DB())

	// ============================================================
	// AUXILIARY ROUTE GROUPS
	// ============================================================
	registerFlowAPI(s)
	s.registerImportListRoute()
	s.registerNetworkRoutes()
	s.registerDBRoutes()
	s.registerVersionRoutes()
	s.registerMirrorSpinRoutes()
	s.registerFileRoutes()
	// s.registerPolicyRoutes() // FIXME: disabled during pgx migration
	s.registerPolicyFileRoutes()
	// s.registerSchemaRoutes() // FIXME: disabled during pgx migration

}

// handleListDomains returns all domains as JSON
func (s *Server) handleListDomains(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	rows, err := s.DB().Query(ctx, `SELECT id, parent_id, name, data, created_at, updated_at FROM domains ORDER BY created_at ASC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Domain struct {
		ID        string          `json:"id"`
		ParentID  sql.NullString  `json:"parent_id"`
		Name      string          `json:"name"`
		Data      json.RawMessage `json:"data"`
		CreatedAt time.Time       `json:"created_at"`
		UpdatedAt time.Time       `json:"updated_at"`
	}

	var result []map[string]any
	for rows.Next() {
		var d Domain
		if err := rows.Scan(&d.ID, &d.ParentID, &d.Name, &d.Data, &d.CreatedAt, &d.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// decode data JSONB
		var payload map[string]any
		_ = json.Unmarshal(d.Data, &payload)

		result = append(result, map[string]any{
			"id":         d.ID,
			"parent_id":  d.ParentID.String, // safe even if NULL
			"name":       d.Name,
			"data":       payload,
			"created_at": d.CreatedAt,
			"updated_at": d.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
