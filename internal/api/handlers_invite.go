package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/core/identity"
)

// handleInviteAccept is implemented in Phase K-1.
// This stub exists so the server builds before Copilot fills it in.
func (s *Server) handleInviteAccept(w http.ResponseWriter, r *http.Request) {

	// Parse body
	var body struct {
		Token     string `json:"token"`
		WhoAreYou string `json:"who_are_you"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Token == "" {
		http.Error(w, "token required", http.StatusBadRequest)
		return
	}
	if body.WhoAreYou == "" {
		http.Error(w, "who_are_you required", http.StatusBadRequest)
		return
	}

	// Ensure DB is configured
	db := s.requireDB(w)
	if db == nil {
		return
	}

	// Begin a DB transaction for the atomic orchestration
	ctx := r.Context()
	tx, err := db.Begin(ctx)
	if err != nil {
		http.Error(w, "failed to begin transaction", http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Resolve the handshake FOR UPDATE to lock the invite row while we operate.
	var subject string
	if err := tx.QueryRow(ctx, `SELECT subject FROM handshakes WHERE token = $1 FOR UPDATE`, body.Token).Scan(&subject); err != nil {
		http.Error(w, "invalid or unknown invite token", http.StatusBadRequest)
		return
	}

	// Basic validation: ensure resolved subject is present
	if subject == "" {
		http.Error(w, "invite subject invalid", http.StatusBadRequest)
		return
	}

	// Call orchestration (thin orchestrator that expects a tx). We pass the
	// invite token; the orchestration will create identity/domain/seat and
	// record receipts. Because we locked the handshake row above (FOR UPDATE),
	// concurrent accept attempts will be serialized by the DB.
	res, err := identity.KnowThyselfAtomic(ctx, tx, body.Token, body.WhoAreYou)
	if err != nil {
		http.Error(w, "failed to accept invite: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with created domain id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"actor_id":   res.ActorID.String(),
		"domain_id":  res.DomainID.String(),
		"receipt_id": res.ReceiptID.String(),
	})
}
