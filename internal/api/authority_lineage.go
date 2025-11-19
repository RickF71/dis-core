package api

import (
	"context"
	"encoding/json"
	"net/http"

	"dis-core/internal/authority"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GOV-9: Authority Lineage API
// Provides read-only access to governance receipt chains with integrity verification

// handleAuthorityLineage handles GET /api/authority/lineage/{domain_id}
func (s *Server) handleAuthorityLineage(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract domain_id from URL
	domainIDStr := chi.URLParam(r, "domain_id")
	if domainIDStr == "" {
		http.Error(w, "domain_id parameter required", http.StatusBadRequest)
		return
	}

	// Parse UUID
	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		http.Error(w, "invalid domain_id format (expected UUID)", http.StatusBadRequest)
		return
	}

	// Get lineage from database
	if s.db == nil {
		http.Error(w, "database unavailable", http.StatusInternalServerError)
		return
	}

	lineage, err := authority.GetAuthorityLineage(ctx, s.db, domainID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lineage)
}

// handleAuthorityLineageVerify handles GET /api/authority/lineage/{domain_id}/verify
// Returns only integrity status without full receipt payloads
func (s *Server) handleAuthorityLineageVerify(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract domain_id from URL
	domainIDStr := chi.URLParam(r, "domain_id")
	if domainIDStr == "" {
		http.Error(w, "domain_id parameter required", http.StatusBadRequest)
		return
	}

	// Parse UUID
	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		http.Error(w, "invalid domain_id format (expected UUID)", http.StatusBadRequest)
		return
	}

	// Get lineage from database
	if s.db == nil {
		http.Error(w, "database unavailable", http.StatusInternalServerError)
		return
	}

	lineage, err := authority.GetAuthorityLineage(ctx, s.db, domainID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return only integrity status
	response := map[string]interface{}{
		"status":    lineage.Status,
		"domain_id": lineage.DomainID,
		"count":     len(lineage.Receipts),
		"integrity": lineage.Integrity,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
