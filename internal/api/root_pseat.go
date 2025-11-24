package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"dis-core/internal/auth"
	"dis-core/internal/contextx"
	"dis-core/internal/db"

	"github.com/jackc/pgx/v5"
)

// handleClaimRootPSeat allows the bootstrap human to claim the null domain's prime seat.
func (s *Server) handleClaimRootPSeat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user := auth.GetActiveUser(r)
	if user == nil || !user.IsBound() {
		JSONError(w, http.StatusUnauthorized, "must be an authenticated bound user to claim root pseat")
		return
	}

	pool := s.requireDB(w)
	if pool == nil {
		return
	}

	// Resolve actor identity id from presentation_name / corporeal UID
	var actorID string
	if err := pool.QueryRow(ctx, `SELECT id::text FROM identities WHERE presentation_name = $1 LIMIT 1`, user.CorporealDomainUID).Scan(&actorID); err != nil {
		JSONError(w, http.StatusBadRequest, fmt.Sprintf("cannot resolve actor identity: %v", err))
		return
	}

	// Fetch bootstrap actor id for policy input
	bootstrapActorID, err := db.GetBootstrapActorID(ctx, pool)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to read bootstrap state: %v", err))
		return
	}

	// Build policy input and evaluate
	if s.policy != nil {
		input := map[string]interface{}{
			"action":    "root_pseat_claim",
			"domain_id": "null",
			"actor_id":  actorID,
			"bootstrap": map[string]interface{}{"actor_id": bootstrapActorID},
		}
		decision, err := s.policy.EvaluateAction(ctx, input)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, fmt.Sprintf("policy eval error: %v", err))
			return
		}
		if decision == nil || !decision.Allow {
			JSONError(w, http.StatusForbidden, "denied by policy")
			return
		}
		// Attach decision details into context so receipts include policy panel
		if decision.Details != nil {
			ctx = contextx.WithPolicyDecisionMap(ctx, decision.Details)
		}
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, fmt.Sprintf("begin tx: %v", err))
		return
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := db.ClaimRootPSeatTx(ctx, tx, actorID); err != nil {
		// Map common errors to user-friendly codes
		if err == pgx.ErrNoRows {
			JSONError(w, http.StatusConflict, "root pseat already occupied")
			return
		}
		JSONError(w, http.StatusBadRequest, fmt.Sprintf("claim failed: %v", err))
		return
	}

	if err := tx.Commit(ctx); err != nil {
		JSONError(w, http.StatusInternalServerError, fmt.Sprintf("commit: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
