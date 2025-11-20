package api

import (
	"testing"
)

// NOTE: Middleware-path end-to-end tests are intentionally skipped for MX-4.x
// These tests should exercise the full middleware resolution (X-External-User
// header -> identity_bindings -> ActiveUser binding) and the router stack.

func TestFreezeDomain_MiddlewarePath_SKIPPED(t *testing.T) {
	t.Skip("MX-4.1: Enable after identity_bindings + dis-bridge tokens are stable")

	// Boilerplate for future:
	// 1. seed domain
	// 2. seed identity binding row
	// 3. set X-External-User header
	// 4. router.ServeHTTP
	// 5. assert created_by == bound actor

	// Example (left as a guide):
	// s, pool := buildTestServer(t)
	// defer pool.Close()
	// domainID := "4daf928e-e58c-454e-8395-f3dedd103dde"
	// actorUID := "489a4841-fdd0-4ef5-a178-46e8cd7612af"
	// // insert identity_bindings row mapping external -> actorUID
	// body := map[string]any{"scope": "all"}
	// b, _ := json.Marshal(body)
	// rr := httptest.NewRecorder()
	// req := httptest.NewRequest(http.MethodPost, "/api/domain/"+domainID+"/freeze", bytes.NewReader(b))
	// req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("X-External-User", "test-external-user-1")
	// s.Router().ServeHTTP(rr, req)
	// require.Equal(t, http.StatusOK, rr.Code)
	// frs, err := db.GetActiveFreezes(req.Context(), pool, domainID)
	// require.NoError(t, err)
	// require.GreaterOrEqual(t, len(frs), 1)
	// require.Equal(t, actorUID, frs[0].CreatedBy)
}
