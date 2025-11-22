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

func TestBootstrapFlow_MinimalNullPseat(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	ctx := context.Background()

	// Ensure a truly empty system for this test by truncating canonical fixtures
	// created by the test harness.
	_, err := pool.Exec(ctx, `TRUNCATE TABLE domain_seats, domains, identities, handshakes, receipts RESTART IDENTITY CASCADE;`)
	require.NoError(t, err)

	// 1) Request establish on an empty system -> expect a challenge
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login/establish", bytes.NewReader([]byte(`{}`)))
	s.Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Logf("establish response body: %s", rr.Body.String())
	}
	require.Equal(t, http.StatusOK, rr.Code)

	var chResp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &chResp)
	require.NoError(t, err)
	require.Equal(t, "challenge", chResp["status"])
	token := chResp["challenge_id"]
	require.NotEmpty(t, token)

	// 2) Accept the invite using the challenge token and a display name
	body := map[string]any{"token": token, "who_are_you": "Rick Bootstrap"}
	b, _ := json.Marshal(body)
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/invite/accept", bytes.NewReader(b))
	s.Router().ServeHTTP(rr2, req2)
	require.Equal(t, http.StatusOK, rr2.Code)

	var resp map[string]string
	err = json.Unmarshal(rr2.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotEmpty(t, resp["actor_id"])
	require.NotEmpty(t, resp["domain_id"])

	// 3) Verify domain.null exists
	var dcount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM domains WHERE name = 'domain.null'`).Scan(&dcount)
	require.NoError(t, err)
	require.Equal(t, 1, dcount)

	// 4) Verify prime seat exists and is assigned to actor
	var seatCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM domain_seats WHERE domain_id = $1::uuid AND member_id = $2 AND seat_type = 'prime'`, resp["domain_id"], resp["actor_id"]).Scan(&seatCount)
	require.NoError(t, err)
	require.GreaterOrEqual(t, seatCount, 1)
}
