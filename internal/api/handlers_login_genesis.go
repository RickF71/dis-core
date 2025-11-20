package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/core/identity"
)

// handleLoginGenesis handles the first human login flow via the K-1 orchestration.
func (s *Server) handleLoginGenesis(w http.ResponseWriter, r *http.Request) {
	// Parse body
	var body struct {
		InviteToken      string `json:"invite_token"`
		PresentationName string `json:"presentation_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.InviteToken == "" {
		http.Error(w, "invite_token required", http.StatusBadRequest)
		return
	}
	if body.PresentationName == "" {
		http.Error(w, "presentation_name required", http.StatusBadRequest)
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

	res, err := identity.KnowThyselfAtomic(ctx, tx, body.InviteToken, body.PresentationName)
	if err != nil {
		http.Error(w, "failed to complete genesis login: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := LoginGenesisResponse{
		Status:           "ok",
		ActorID:          res.ActorID.String(),
		DomainID:         res.DomainID.String(),
		ReceiptID:        res.ReceiptID.String(),
		PresentationName: body.PresentationName,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
