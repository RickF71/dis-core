package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/core/identity"
)

// handleLoginSession initializes a read-only actor session and returns an ActorContext.
func (s *Server) handleLoginSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ActorID  string `json:"actor_id"`
		DomainID string `json:"domain_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.ActorID == "" || body.DomainID == "" {
		http.Error(w, "actor_id and domain_id are required", http.StatusBadRequest)
		return
	}

	db := s.requireDB(w)
	if db == nil {
		return
	}

	ctx := r.Context()
	tx, err := db.Begin(ctx)
	if err != nil {
		http.Error(w, "failed to begin transaction", http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	actx, err := identity.LoadActorContextTx(ctx, tx, body.ActorID, body.DomainID)
	if err != nil {
		http.Error(w, "failed to load actor context: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actx)
}
