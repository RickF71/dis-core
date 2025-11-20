package api

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
)

func TestLoginEstablish_CreatesSessionToken(t *testing.T) {
    s, pool := buildTestServer(t)
    defer pool.Close()

    ctx := context.Background()
    token := "genesis-token-establish"
    subject := "human:zoe"
    id := "handshake-3"
    // seed handshake
    _, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id, token, subject)
    require.NoError(t, err)

    // perform genesis to create actor/domain/seat
    body := map[string]any{"invite_token": token, "presentation_name": "Zoe Example"}
    b, _ := json.Marshal(body)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodPost, "/api/login/genesis", bytes.NewReader(b))
    s.Router().ServeHTTP(rr, req)
    require.Equal(t, http.StatusOK, rr.Code)

    var resp LoginGenesisResponse
    err = json.Unmarshal(rr.Body.Bytes(), &resp)
    require.NoError(t, err)

    // call /api/login/establish to create session
    eb := map[string]any{"actor_id": resp.ActorID, "domain_id": resp.DomainID}
    ebj, _ := json.Marshal(eb)
    errrec := httptest.NewRecorder()
    ereq := httptest.NewRequest(http.MethodPost, "/api/login/establish", bytes.NewReader(ebj))
    s.Router().ServeHTTP(errrec, ereq)
    require.Equal(t, http.StatusOK, errrec.Code)

    var eres map[string]any
    err = json.Unmarshal(errrec.Body.Bytes(), &eres)
    require.NoError(t, err)
    require.Equal(t, "ok", eres["status"])
    tok, ok := eres["token"].(string)
    require.True(t, ok)
    require.NotEmpty(t, tok)

    // verify session row exists and expiry approx 8h
    var created time.Time
    var expires time.Time
    err = pool.QueryRow(ctx, `SELECT created_at, expires_at FROM sessions WHERE token = $1`, tok).Scan(&created, &expires)
    require.NoError(t, err)
    // expires should be roughly created + 8h (give 5s leeway)
    delta := expires.Sub(created)
    require.GreaterOrEqual(t, delta, 8*time.Hour)
    require.Less(t, delta, 8*time.Hour+5*time.Second)
}
