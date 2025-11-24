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

func TestLoginWithID_CreatesSession(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	ctx := context.Background()

	// Insert an identity (corporeal) and attach a prime seat pointing to a
	// known corporeal domain seeded by the test harness.
	var identityID string
	err := pool.QueryRow(ctx, `INSERT INTO identities (id, subject, presentation_name, identity_type) VALUES (gen_random_uuid(), $1, $2, 'corporeal') RETURNING id::text`, "human:test", "Test User").Scan(&identityID)
	require.NoError(t, err)

	// Use the canonical seeded corporeal domain from the test bootstrap
	domainID := "a1111111-1111-1111-1111-111111111111"

	_, err = pool.Exec(ctx, `INSERT INTO domain_seats (id, domain_id, member_id, seat_type) VALUES (gen_random_uuid(), $1, $2, 'prime')`, domainID, identityID)
	require.NoError(t, err)

	// Call the login-with-id endpoint
	body := map[string]any{"identity_id": identityID}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login/with-id", bytes.NewReader(b))

	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]any
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	tok, ok := resp["session_token"].(string)
	require.True(t, ok, "session_token should be present")
	require.NotEmpty(t, tok, "session_token should not be empty")

	// Verify session row exists for returned token
	var got string
	err = pool.QueryRow(ctx, `SELECT token FROM sessions WHERE token = $1`, tok).Scan(&got)
	require.NoError(t, err)
	require.Equal(t, tok, got)
}
