package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"dis-core/internal/core/identity"
	"dis-core/internal/db"
	"dis-core/internal/util"

	"github.com/jackc/pgx/v5"
)

// handleInviteAccept is implemented in Phase K-1.
// This stub exists so the server builds before Copilot fills it in.
func (s *Server) handleInviteAccept(w http.ResponseWriter, r *http.Request) {
	// Parse body
	var body struct {
		Token            string `json:"token"`
		PresentationName string `json:"presentation_name,omitempty"`
		// legacy field used by some tests/clients
		WhoAreYou string `json:"who_are_you,omitempty"`
		// Optional: callers may supply a domain_id to indicate intent. If
		// provided we enforce UID-only semantics.
		DomainID string `json:"domain_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Token == "" {
		http.Error(w, "token required", http.StatusBadRequest)
		return
	}

	// Allow legacy callers to supply who_are_you; prefer explicit presentation_name.
	if body.PresentationName == "" && body.WhoAreYou != "" {
		body.PresentationName = body.WhoAreYou
	}

	// If a domain_id was provided, validate it is UID-only (UUID string).
	if body.DomainID != "" {
		if err := util.ValidateDomainUID(body.DomainID); err != nil {
			http.Error(w, "invalid domain_id: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Ensure DB is configured
	pool := s.requireDB(w)
	if pool == nil {
		return
	}

	ctx := r.Context()
	var res *identity.KnowThyselfResult
	err := db.WithTx(ctx, pool, func(tx pgx.Tx) error {
		var err error
		res, err = identity.KnowThyselfAtomic(ctx, tx, body.Token, body.PresentationName)
		return err
	})
	if err != nil {
		if errors.Is(err, identity.ErrUnknownInvite) {
			http.Error(w, "invalid or unknown invite token", http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to accept invite: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with created ids
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"identity_id": res.IdentityID.String(),
		"domain_id":   res.DomainID.String(),
		"actor_id":    res.ActorID.String(),
		"receipt_id":  res.ReceiptID.String(),
	})
}
