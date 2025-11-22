package api

import (
	"encoding/json"
	"net/http"
	"time"

	"dis-core/internal/core/identity"
	"dis-core/internal/core/session"
	"dis-core/internal/util"

	"github.com/google/uuid"
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
	// Note: if the system is empty (no identities and no domains) we
	// support a minimal bootstrap path: create a one-time handshake token
	// marked as a genesis challenge and return it. The client can then
	// submit that token to /api/invite/accept to complete the bootstrap.

	db := s.requireDB(w)
	if db == nil {
		return
	}

	ctx := r.Context()

	// Quick check for empty system: if both identities and domains are zero
	// we emit a single-use genesis handshake token and return a challenge.
	var idCount int64
	var domainCount int64
	if err := db.QueryRow(ctx, `SELECT COUNT(1) FROM identities`).Scan(&idCount); err == nil {
		_ = db.QueryRow(ctx, `SELECT COUNT(1) FROM domains`).Scan(&domainCount)
	}
	if idCount == 0 && domainCount == 0 {
		// generate handshake token
		token := uuid.NewString()
		hid := uuid.NewString()
		if _, err := db.Exec(ctx, `INSERT INTO handshakes (id, token, subject, status, created_at) VALUES ($1,$2,$3,$4,NOW())`, hid, token, "", "genesis"); err != nil {
			http.Error(w, "failed to create bootstrap challenge: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "challenge", "challenge_id": token})
		return
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		http.Error(w, "failed to begin transaction", http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Validate that DomainID is a UID-only value before loading context
	if err := util.ValidateDomainUID(body.DomainID); err != nil {
		http.Error(w, "invalid domain_id: "+err.Error(), http.StatusBadRequest)
		return
	}

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
