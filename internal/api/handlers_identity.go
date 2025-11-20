package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/identity"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// handleGetIdentityLineage returns the full identity lineage for a given actor_id,
// including all receipts and hash chain integrity verification.
func (s *Server) handleGetIdentityLineage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	actorIDStr := chi.URLParam(r, "actor_id")
	actorID, err := uuid.Parse(actorIDStr)
	if err != nil {
		http.Error(w, "invalid actor_id format", http.StatusBadRequest)
		return
	}

	db := s.requireDB(w)
	if db == nil {
		return
	}

	lineage, err := identity.GetIdentityLineage(ctx, db, actorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(lineage); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleGetIdentityLineageVerify returns only the integrity status summary
// for a given actor_id, without the full payload data.
func (s *Server) handleGetIdentityLineageVerify(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	actorIDStr := chi.URLParam(r, "actor_id")
	actorID, err := uuid.Parse(actorIDStr)
	if err != nil {
		http.Error(w, "invalid actor_id format", http.StatusBadRequest)
		return
	}

	db := s.requireDB(w)
	if db == nil {
		return
	}

	lineage, err := identity.GetIdentityLineageSummary(ctx, db, actorID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":    "ok",
		"actor_id":  actorID.String(),
		"integrity": lineage.Integrity,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
