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

func TestLoginGenesis_InvalidDomainID_Returns400(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	ctx := context.Background()
	token := "genesis-token-xyz"
	subject := "human:eva"
	id := "handshake-1"
	// seed handshake
	_, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id, token, subject)
	require.NoError(t, err)

	// provide invalid domain_id
	body := map[string]any{"invite_token": token, "presentation_name": "Eva Example", "domain_id": "not-a-uuid"}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login/genesis", bytes.NewReader(b))

	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestInviteAccept_InvalidDomainID_Returns400(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	ctx := context.Background()
	token := "invite-token-123"
	subject := "human:alice"
	id := "handshake-invite-1"
	// seed handshake
	_, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id, token, subject)
	require.NoError(t, err)

	body := map[string]any{"token": token, "who_are_you": "Alice Example", "domain_id": "bad-uuid"}
	b, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/invite/accept", bytes.NewReader(b))

	s.Router().ServeHTTP(rr, req)
	require.Equal(t, http.StatusBadRequest, rr.Code)
}
