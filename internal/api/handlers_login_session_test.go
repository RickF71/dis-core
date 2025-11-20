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

func TestLoginSession_ReadOnlyPrimeSeatContext(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	ctx := context.Background()
	token := "genesis-token-session"
	subject := "human:alice"
	id := "handshake-2"
	// seed handshake
	_, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id, token, subject)
	require.NoError(t, err)

	// perform genesis to create actor/domain/seat
	body := map[string]any{"invite_token": token, "presentation_name": "Alice Example"}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login/genesis", bytes.NewReader(b))
	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var resp LoginGenesisResponse
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	// capture current seat count for domain+actor (read-only assertion later)
	var beforeSeatCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM domain_seats WHERE domain_id = $1::uuid AND member_id = $2`, resp.DomainID, resp.ActorID).Scan(&beforeSeatCount)
	require.NoError(t, err)

	// Now call /api/login/session
	sessionBody := map[string]any{"actor_id": resp.ActorID, "domain_id": resp.DomainID}
	sb, _ := json.Marshal(sessionBody)
	srr := httptest.NewRecorder()
	sreq := httptest.NewRequest(http.MethodPost, "/api/login/session", bytes.NewReader(sb))
	s.Router().ServeHTTP(srr, sreq)
	require.Equal(t, http.StatusOK, srr.Code)

	var actx map[string]any
	err = json.Unmarshal(srr.Body.Bytes(), &actx)
	require.NoError(t, err)

	// Assertions
	// seat_id present
	require.NotEmpty(t, actx["seat_id"])
	// presentation_name matches
	require.Equal(t, "Alice Example", actx["presentation_name"])
	// domain_fdn resolves to domain name
	var domainName string
	err = pool.QueryRow(ctx, `SELECT name FROM domains WHERE id = $1::uuid`, resp.DomainID).Scan(&domainName)
	require.NoError(t, err)
	require.Equal(t, domainName, actx["domain_fdn"])

	// Ensure no mutation occurred: seat count unchanged
	var afterSeatCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM domain_seats WHERE domain_id = $1::uuid AND member_id = $2`, resp.DomainID, resp.ActorID).Scan(&afterSeatCount)
	require.NoError(t, err)
	require.Equal(t, beforeSeatCount, afterSeatCount)
}
