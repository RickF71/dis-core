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

func TestLoginGenesis_FullFlow(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	ctx := context.Background()
	token := "genesis-token-xyz"
	subject := "human:eva"
	id := "handshake-1"
	// seed handshake
	_, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id, token, subject)
	require.NoError(t, err)

	// perform POST /api/login/genesis
	body := map[string]any{"invite_token": token, "presentation_name": "Eva Example"}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login/genesis", bytes.NewReader(b))

	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var resp LoginGenesisResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, "ok", resp.Status)
	require.NotEmpty(t, resp.ActorID)
	require.NotEmpty(t, resp.DomainID)
	require.NotEmpty(t, resp.ReceiptID)
	require.Equal(t, "Eva Example", resp.PresentationName)

	// verify handshake no longer exists (consumed)
	var count int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM handshakes WHERE token = $1`, token).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	// verify identity row exists with presentation_name
	var gotPresentation string
	err = pool.QueryRow(ctx, `SELECT presentation_name FROM identities WHERE presentation_name = $1 LIMIT 1`, "Eva Example").Scan(&gotPresentation)
	require.NoError(t, err)
	require.Equal(t, "Eva Example", gotPresentation)

	// verify prime seat exists for the created domain and actor
	var seatCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM domain_seats WHERE domain_id = $1::uuid AND member_id = $2 AND seat_type = 'prime'`, resp.DomainID, resp.ActorID).Scan(&seatCount)
	require.NoError(t, err)
	require.GreaterOrEqual(t, seatCount, 1)

	// verify receipt exists and is ci.call.login.v1
	var rtype string
	err = pool.QueryRow(ctx, `SELECT type FROM receipts WHERE id = $1::uuid`, resp.ReceiptID).Scan(&rtype)
	require.NoError(t, err)
	require.Equal(t, "ci.call.login.v1", rtype)
}
