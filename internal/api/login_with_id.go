package api

import (
	"encoding/json"
	"net/http"
	"time"

	"dis-core/internal/core/session"
)

type loginWithIDRequest struct {
	IdentityID string `json:"identity_id"`
}

type loginWithIDResponse struct {
	SessionToken      string `json:"session_token"`
	IdentityID        string `json:"identity_id"`
	CorporealDomainID string `json:"corporeal_domain_id"`
	ActorID           string `json:"actor_id"`
}

func (s *Server) handleLoginWithID(w http.ResponseWriter, r *http.Request) {
	var req loginWithIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid_json", http.StatusBadRequest)
		return
	}

	if req.IdentityID == "" {
		http.Error(w, "missing_identity_id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	pool := s.DB()
	if pool == nil {
		http.Error(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	// Verify identity exists
	var identityID string
	if err := pool.QueryRow(ctx, `SELECT id::text FROM identities WHERE id = $1::uuid`, req.IdentityID).Scan(&identityID); err != nil {
		http.Error(w, "identity_not_found", http.StatusNotFound)
		return
	}

	// Resolve a seat and domain for this identity. We expect a domain_seats
	// row with member_id matching the identity (created by bootstrap/genesis
	// flows). If none found, return not found / not bound.
	var seatID string
	var corporealDomainID string
	if err := pool.QueryRow(ctx, `SELECT id::text, domain_id::text FROM domain_seats WHERE member_id = $1 LIMIT 1`, identityID).Scan(&seatID, &corporealDomainID); err != nil {
		http.Error(w, "identity_not_bound", http.StatusNotFound)
		return
	}

	// Create a new session token inside a transaction using CreateSessionTx
	tx, err := pool.Begin(ctx)
	if err != nil {
		http.Error(w, "session_tx_begin_failed", http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Use 7 day TTL for this programmatic session. Use the identity as the
	// actor id and the resolved seat/domain from domain_seats.
	_, token, err := session.CreateSessionTx(ctx, tx, identityID, corporealDomainID, seatID, 7*24*time.Hour)
	if err != nil {
		http.Error(w, "session_creation_failed", http.StatusInternalServerError)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		http.Error(w, "session_commit_failed", http.StatusInternalServerError)
		return
	}

	out := loginWithIDResponse{
		SessionToken:      token,
		IdentityID:        identityID,
		CorporealDomainID: corporealDomainID,
		ActorID:           identityID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}
