package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test that a valid persistent session token (created via /api/login/establish)
// allows access to /api/me and returns the expected bound domain info.
func TestMe_WithValidSession(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	ctx := context.Background()
	token := "invite-token-me-1"
	subject := "human:me"
	id := "handshake-me-1"
	// seed handshake
	_, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id, token, subject)
	require.NoError(t, err)

	// perform genesis to create actor/domain/seat
	body := map[string]any{"invite_token": token, "presentation_name": "Me Example"}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login/genesis", bytes.NewReader(b))
	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var gresp LoginGenesisResponse
	err = json.Unmarshal(rr.Body.Bytes(), &gresp)
	require.NoError(t, err)

	// call /api/login/establish to create session
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

	// Use token to call GET /api/me
	mr := httptest.NewRecorder()
	mreq := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	mreq.Header.Set("Authorization", "Bearer "+tok)
	s.Router().ServeHTTP(mr, mreq)
	require.Equal(t, http.StatusOK, mr.Code)

	var mresp MeResponse
	err = json.Unmarshal(mr.Body.Bytes(), &mresp)
	require.NoError(t, err)
	// The middleware binds the session's domain id into CorporealDomainUID
	require.True(t, mresp.Bound)
	require.Equal(t, gresp.DomainID, mresp.CorporealDomainUID)
}

// Invalid or nonexistent session token must not authenticate the request.
func TestMe_InvalidSessionToken(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	// Call /api/me with a bogus token
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token-xyz")
	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}
