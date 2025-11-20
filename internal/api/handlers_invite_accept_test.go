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

func TestInviteAccept_HappyPath(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	ctx := context.Background()
	token := "invite-token-123"
	subject := "human:alice"
	id := "handshake-invite-1"
	// seed handshake
	_, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id, token, subject)
	require.NoError(t, err)

	// perform POST /api/invite/accept
	body := map[string]any{"token": token, "who_are_you": "Alice Example"}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/invite/accept", bytes.NewReader(b))

	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotEmpty(t, resp["actor_id"])
	require.NotEmpty(t, resp["domain_id"])
	require.NotEmpty(t, resp["receipt_id"])

	// verify handshake consumed
	var count int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM handshakes WHERE token = $1`, token).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	// verify identity row exists with presentation_name
	var gotPresentation string
	err = pool.QueryRow(ctx, `SELECT presentation_name FROM identities WHERE presentation_name = $1 LIMIT 1`, "Alice Example").Scan(&gotPresentation)
	require.NoError(t, err)
	require.Equal(t, "Alice Example", gotPresentation)

	// verify prime seat exists for created domain and actor
	var seatCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM domain_seats WHERE domain_id = $1::uuid AND member_id = $2 AND seat_type = 'prime'`, resp["domain_id"], resp["actor_id"]).Scan(&seatCount)
	require.NoError(t, err)
	require.GreaterOrEqual(t, seatCount, 1)

	// verify receipt exists and is ci.call.login.v1
	var rtype string
	err = pool.QueryRow(ctx, `SELECT type FROM receipts WHERE id = $1::uuid`, resp["receipt_id"]).Scan(&rtype)
	require.NoError(t, err)
	require.Equal(t, "ci.call.login.v1", rtype)
}

func TestInviteAccept_InvalidToken(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	// no handshake seeded for this token
	token := "nonexistent-token-999"

	body := map[string]any{"token": token, "who_are_you": "Nobody"}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/invite/accept", bytes.NewReader(b))

	s.Router().ServeHTTP(rr, req)
	// handler treats KnowThyselfAtomic errors as BadRequest
	require.Equal(t, http.StatusBadRequest, rr.Code)

	// ensure no identity was created for the presentation name
	var count int
	err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM identities WHERE presentation_name = $1`, "Nobody").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}
