package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/auth"
	dbstore "dis-core/internal/db"

	"github.com/go-chi/chi/v5"
)

// handleAddSeatRole handles POST /api/seat/{id}/roles/add
// Body: { "role": "owner" }
func (s *Server) handleAddSeatRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := s.requireDB(w)
	if db == nil {
		return
	}
	ctx := r.Context()

	user := auth.GetActiveUser(r)
	if user == nil || !user.IsAuthenticated() {
		JSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	activeSeatID, hasActive := auth.GetActiveActor(r)
	if !hasActive || activeSeatID == "" {
		JSONError(w, http.StatusUnauthorized, "active actor required")
		return
	}

	targetSeatID := chi.URLParam(r, "id")
	if targetSeatID == "" {
		JSONError(w, http.StatusBadRequest, "seat id required")
		return
	}

	var body struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.Role == "" {
		JSONError(w, http.StatusBadRequest, "role required")
		return
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to begin tx")
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Load caller roles for active seat
	callerRoles, err := dbstore.ListRolesForSeatTx(ctx, tx, activeSeatID)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to load caller roles")
		return
	}

	// Require 'owner' role to assign roles
	isOwner := false
	for _, r := range callerRoles {
		if r == "owner" {
			isOwner = true
			break
		}
	}
	if !isOwner {
		JSONError(w, http.StatusForbidden, "requires owner role to assign roles")
		return
	}

	if err := dbstore.AddRoleTx(ctx, tx, targetSeatID, body.Role); err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to add role")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to commit")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleListSeatRoles handles GET /api/seat/{id}/roles
func (s *Server) handleListSeatRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := s.requireDB(w)
	if db == nil {
		return
	}
	ctx := r.Context()

	targetSeatID := chi.URLParam(r, "id")
	if targetSeatID == "" {
		JSONError(w, http.StatusBadRequest, "seat id required")
		return
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to begin tx")
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	roles, err := dbstore.ListRolesForSeatTx(ctx, tx, targetSeatID)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to list roles")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to commit")
		return
	}

	JSON(w, http.StatusOK, map[string]any{"roles": roles})
}
