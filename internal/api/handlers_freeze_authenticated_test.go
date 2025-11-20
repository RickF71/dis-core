// NOTE: This test exercises the *authenticated path*.
// ActiveUser is injected manually to avoid middleware overwriting.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dis-core/internal/db"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestFreezeDomain_UsesAuthenticatedActor verifies that when an ActiveUser
// is bound in the request context, the server uses that actor as created_by
// for domain freeze operations.
func TestFreezeDomain_UsesAuthenticatedActor(t *testing.T) {
	s, pool := buildTestServer(t)
	defer pool.Close()

	// pick a seeded test domain (testdb provides a deterministic domain id)
	domainID := "4daf928e-e58c-454e-8395-f3dedd103dde"

	// Actor who performs the freeze - must be a UUID string since server parses it
	actorUID := uuid.NewString()

	// Build request WITHOUT created_by in body so handler will prefer ActiveUser
	// include a valid scope; TTL must be an RFC3339 string if present so omit it here
	body := map[string]any{"scope": "all"}
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/domain/"+domainID+"/freeze", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")

	// Attach an authenticated ActiveUser
	// Attach ActiveUser directly to the request context so the handler will
	// prefer this authenticated actor when selecting created_by.
	req = WithTestActiveUser(req, actorUID)

	rr := httptest.NewRecorder()

	// Call the handler directly to avoid the middleware stack (which in tests
	// may attempt DB-backed resolution). We need to set chi's route context so
	// chi.URLParam() inside the handler can find the domain id.
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", domainID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

	s.handleFreezeDomain(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	// Validate freeze record stored with CreatedBy = actorUID
	frs, err := db.GetActiveFreezes(req.Context(), pool, domainID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(frs), 1)
	// CreatedBy is stored as UUID string in freeze record
	require.Equal(t, actorUID, frs[0].CreatedBy)
}
