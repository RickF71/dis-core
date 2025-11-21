package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"dis-core/internal/auth"
	"dis-core/internal/core/session"
)

// POST /api/logout
// Revoke the current session token (marks revoked_at = now()).
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	user := auth.GetActiveUser(r)
	if user == nil {
		JSONError(w, http.StatusUnauthorized, "No active user context")
		return
	}

	// Extract bearer token from Authorization header
	authz := strings.TrimSpace(r.Header.Get("Authorization"))
	if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
		JSONError(w, http.StatusUnauthorized, "missing bearer token")
		return
	}
	token := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
	if token == "" {
		JSONError(w, http.StatusUnauthorized, "missing bearer token")
		return
	}

	db := s.DB()
	if db == nil {
		JSONError(w, http.StatusServiceUnavailable, "database not available")
		return
	}

	ctx := r.Context()
	tx, err := db.Begin(ctx)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "could not begin tx")
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Revoke the session by token
	if err := session.RevokeSessionByTokenTx(ctx, tx, token); err != nil {
		JSONError(w, http.StatusInternalServerError, "could not revoke session")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		JSONError(w, http.StatusInternalServerError, "could not commit revoke")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/sessions
// List active sessions for the authenticated user (non-revoked, non-expired)
func (s *Server) handleListSessions(w http.ResponseWriter, r *http.Request) {
	user := auth.GetActiveUser(r)
	if user == nil {
		JSONError(w, http.StatusUnauthorized, "No active user context")
		return
	}

	// Get active seat ID from context
	seatID, ok := auth.GetActiveActor(r)
	if !ok || seatID == "" {
		JSONError(w, http.StatusBadRequest, "no active actor seat")
		return
	}

	db := s.DB()
	if db == nil {
		JSONError(w, http.StatusServiceUnavailable, "database not available")
		return
	}

	ctx := r.Context()
	tx, err := db.Begin(ctx)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "could not begin tx")
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// domain id is stored in ActiveUser.CorporealDomainUID
	domainID := user.CorporealDomainUID

	sessions, err := session.ListActiveSessionsForSeatDomainTx(ctx, tx, seatID, domainID)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "could not list sessions")
		return
	}

	// determine current token from Authorization header
	authz := strings.TrimSpace(r.Header.Get("Authorization"))
	var currentToken string
	if strings.HasPrefix(authz, "Bearer ") {
		currentToken = strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
	}

	type respSession struct {
		ID        string     `json:"id"`
		CreatedAt time.Time  `json:"created_at"`
		ExpiresAt time.Time  `json:"expires_at"`
		RevokedAt *time.Time `json:"revoked_at"`
		Current   bool       `json:"current"`
	}

	out := make([]respSession, 0, len(sessions))
	for _, ss := range sessions {
		out = append(out, respSession{
			ID:        ss.ID,
			CreatedAt: ss.CreatedAt,
			ExpiresAt: ss.ExpiresAt,
			RevokedAt: ss.RevokedAt,
			Current:   currentToken != "" && ss.Token == currentToken,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		JSONError(w, http.StatusInternalServerError, "could not commit tx")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"sessions": out})
}
