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

// Unauthenticated requests should be rejected
func TestWhoAmI_Unauthenticated_Returns401(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/whoami", nil)
	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

// Authenticated session should return domain_id and seat_id (and actor_id when resolvable)
func TestWhoAmI_WithValidSession_Returns200AndJSON(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	ctx := context.Background()
	token := "invite-token-whoami-1"
	subject := "human:whoami"
	id := "handshake-whoami-1"
	// seed handshake
	_, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id, token, subject)
	require.NoError(t, err)

	// perform genesis to create actor/domain/seat
	body := map[string]any{"invite_token": token, "presentation_name": "Who Am I Example"}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login/genesis", bytes.NewReader(b))
	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var gresp LoginGenesisResponse
	err = json.Unmarshal(rr.Body.Bytes(), &gresp)
	require.NoError(t, err)

	// create persistent session via login/establish
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

	// actor_context is returned — use its seat_id for expected value
	var expectedSeat string
	if ac, ok := eres["actor_context"].(map[string]interface{}); ok {
		if sid, ok2 := ac["seat_id"].(string); ok2 {
			expectedSeat = sid
		}
	}

	// Call /api/whoami with token
	mr := httptest.NewRecorder()
	mreq := httptest.NewRequest(http.MethodGet, "/api/whoami", nil)
	mreq.Header.Set("Authorization", "Bearer "+tok)
	s.Router().ServeHTTP(mr, mreq)
	require.Equal(t, http.StatusOK, mr.Code)

	var wresp map[string]string
	err = json.Unmarshal(mr.Body.Bytes(), &wresp)
	require.NoError(t, err)
	require.Equal(t, gresp.DomainID, wresp["domain_id"])
	// seat id should match established actor_context seat id when present
	if expectedSeat != "" {
		require.Equal(t, expectedSeat, wresp["seat_id"])
	}
}
