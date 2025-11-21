package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bytes"

	"github.com/stretchr/testify/require"
)

// Happy path: create two sessions for the same seat/domain and ensure both are returned
// and the requester's token is marked as current.
func TestListSessions_HappyPath(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	ctx := context.Background()
	token := "list-token-1"
	subject := "human:list"
	id := "handshake-list-1"
	_, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id, token, subject)
	require.NoError(t, err)

	// perform genesis
	body := map[string]any{"invite_token": token, "presentation_name": "List Example"}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login/genesis", bytes.NewReader(b))
	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var gresp LoginGenesisResponse
	err = json.Unmarshal(rr.Body.Bytes(), &gresp)
	require.NoError(t, err)

	// establish first session
	eb := map[string]any{"actor_id": gresp.ActorID, "domain_id": gresp.DomainID}
	ebj, _ := json.Marshal(eb)
	errrec := httptest.NewRecorder()
	ereq := httptest.NewRequest(http.MethodPost, "/api/login/establish", bytes.NewReader(ebj))
	s.Router().ServeHTTP(errrec, ereq)
	require.Equal(t, http.StatusOK, errrec.Code)

	var eres map[string]any
	err = json.Unmarshal(errrec.Body.Bytes(), &eres)
	require.NoError(t, err)
	tok1, ok := eres["token"].(string)
	require.True(t, ok)
	require.NotEmpty(t, tok1)

	// establish second session (simulate another client)
	errrec2 := httptest.NewRecorder()
	ereq2 := httptest.NewRequest(http.MethodPost, "/api/login/establish", bytes.NewReader(ebj))
	s.Router().ServeHTTP(errrec2, ereq2)
	require.Equal(t, http.StatusOK, errrec2.Code)
	var eres2 map[string]any
	err = json.Unmarshal(errrec2.Body.Bytes(), &eres2)
	require.NoError(t, err)
	tok2, ok := eres2["token"].(string)
	require.True(t, ok)
	require.NotEmpty(t, tok2)

	// Call GET /api/sessions with tok1
	lr := httptest.NewRecorder()
	lreq := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	lreq.Header.Set("Authorization", "Bearer "+tok1)
	s.Router().ServeHTTP(lr, lreq)
	require.Equal(t, http.StatusOK, lr.Code)

	var listResp map[string][]map[string]any
	err = json.Unmarshal(lr.Body.Bytes(), &listResp)
	require.NoError(t, err)
	sessions, ok := listResp["sessions"]
	require.True(t, ok)
	require.GreaterOrEqual(t, len(sessions), 2)

	// Ensure at least one session is marked current
	foundCurrent := false
	for _, srow := range sessions {
		if cur, _ := srow["current"].(bool); cur {
			foundCurrent = true
			break
		}
	}
	require.True(t, foundCurrent)
}

func TestListSessions_RequiresAuth(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}
