package api

import (
	"encoding/json"
	"net/http"
	"time"

	"dis-core/internal/core/identity"
	"dis-core/internal/core/session"
)

// handleLoginEstablish creates a persistent session token (TTL 8h) for an actor/domain
func (s *Server) handleLoginEstablish(w http.ResponseWriter, r *http.Request) {
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

	// Validate actor context inside the transaction
	actx, err := identity.LoadActorContextTx(ctx, tx, body.ActorID, body.DomainID)
	if err != nil {
		http.Error(w, "failed to load actor context: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create session with TTL 8 hours
	const ttl = 8 * time.Hour
	_, token, err := session.CreateSessionTx(ctx, tx, actx.ActorID, actx.DomainID, actx.SeatID, ttl)
	if err != nil {
		http.Error(w, "failed to create session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]any{"status": "ok", "token": token, "actor_context": actx}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
