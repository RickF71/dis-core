package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bytes"

	"github.com/stretchr/testify/require"
)

// Test that POST /api/logout revokes the current session and subsequent use of the token is rejected.
func TestLogout_RevokesCurrentSession(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	ctx := context.Background()
	token := "logout-token-1"
	subject := "human:logout"
	id := "handshake-logout-1"
	_, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id, token, subject)
	require.NoError(t, err)

	// perform genesis
	body := map[string]any{"invite_token": token, "presentation_name": "Logout Example"}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login/genesis", bytes.NewReader(b))
	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var gresp LoginGenesisResponse
	err = json.Unmarshal(rr.Body.Bytes(), &gresp)
	require.NoError(t, err)

	// establish session
	eb := map[string]any{"actor_id": gresp.ActorID, "domain_id": gresp.DomainID}
	ebj, _ := json.Marshal(eb)
	errrec := httptest.NewRecorder()
	ereq := httptest.NewRequest(http.MethodPost, "/api/login/establish", bytes.NewReader(ebj))
	s.Router().ServeHTTP(errrec, ereq)
	require.Equal(t, http.StatusOK, errrec.Code)

	var eres map[string]any
	err = json.Unmarshal(errrec.Body.Bytes(), &eres)
	require.NoError(t, err)
	tok, ok := eres["token"].(string)
	require.True(t, ok)
	require.NotEmpty(t, tok)

	// Call POST /api/logout with token
	lr := httptest.NewRecorder()
	lreq := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	lreq.Header.Set("Authorization", "Bearer "+tok)
	s.Router().ServeHTTP(lr, lreq)
	require.Equal(t, http.StatusNoContent, lr.Code)

	// Verify DB shows revoked_at set for this token
	var revokedAt time.Time
	err = pool.QueryRow(ctx, `SELECT revoked_at FROM sessions WHERE token = $1`, tok).Scan(&revokedAt)
	require.NoError(t, err)
	require.False(t, revokedAt.IsZero())

	// Subsequent use of token should be rejected
	mr := httptest.NewRecorder()
	mreq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	mreq.Header.Set("Authorization", "Bearer "+tok)
	s.Router().ServeHTTP(mr, mreq)
	require.Equal(t, http.StatusUnauthorized, mr.Code)
}

// Test that a revoked session token is rejected when calling /api/me
func TestMe_UsesRevokedSessionToken_Returns401(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	ctx := context.Background()
	token := "revoked-token-1"
	subject := "human:revoked"
	id := "handshake-revoked-1"
	_, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id, token, subject)
	require.NoError(t, err)

	// perform genesis
	body := map[string]any{"invite_token": token, "presentation_name": "Revoked Example"}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login/genesis", bytes.NewReader(b))
	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var gresp LoginGenesisResponse
	err = json.Unmarshal(rr.Body.Bytes(), &gresp)
	require.NoError(t, err)

	// establish session
	eb := map[string]any{"actor_id": gresp.ActorID, "domain_id": gresp.DomainID}
	ebj, _ := json.Marshal(eb)
	errrec := httptest.NewRecorder()
	ereq := httptest.NewRequest(http.MethodPost, "/api/login/establish", bytes.NewReader(ebj))
	s.Router().ServeHTTP(errrec, ereq)
	require.Equal(t, http.StatusOK, errrec.Code)

	var eres map[string]any
	err = json.Unmarshal(errrec.Body.Bytes(), &eres)
	require.NoError(t, err)
	tok, ok := eres["token"].(string)
	require.True(t, ok)
	require.NotEmpty(t, tok)

	// Directly mark revoked in DB
	_, err = pool.Exec(ctx, `UPDATE sessions SET revoked_at = now() WHERE token = $1`, tok)
	require.NoError(t, err)

	// Use token to call GET /api/me - must be 401
	mr := httptest.NewRecorder()
	mreq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	mreq.Header.Set("Authorization", "Bearer "+tok)
	s.Router().ServeHTTP(mr, mreq)
	require.Equal(t, http.StatusUnauthorized, mr.Code)
}
