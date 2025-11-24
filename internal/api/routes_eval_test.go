package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dis-core/internal/db"
	"dis-core/internal/ledger"
	"dis-core/internal/policy"
	testdb "dis-core/internal/testdb"

	"github.com/stretchr/testify/require"
)

// Test that the /api/eval route enforces policy denies and still persists a
// receipt containing the policy_ref.
func TestEvalRoute_EnforcesDenyAndPersistsPolicyRef(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	// Build policy engine from local modules (use relative paths so tests
	// running in package/internal/api can find the files).
	gm, err := policy.LoadRegoBundle("../policy/gates.rego")
	require.NoError(t, err)
	rm, err := policy.LoadRegoBundle("../policy/risk.rego")
	require.NoError(t, err)
	fm, err := policy.LoadRegoBundle("../policy/freeze.rego")
	require.NoError(t, err)
	modules := map[string]string{
		"gates.rego":  gm,
		"risk.rego":   rm,
		"freeze.rego": fm,
	}
	eng, err := policy.NewEngine(modules)
	require.NoError(t, err)

	// Open ledger using the test pool
	led, err := ledger.Open(context.Background(), "", pool, nil)
	require.NoError(t, err)

	// Create server with policy engine wired
	s := NewWithPolicy(pool, led, eng, nil)
	require.NotNil(t, s)

	// Register eval route explicitly with the engine
	s.RegisterEvalRoute(eng)

	// Build request that should be denied by AT-1 rule
	body := map[string]any{
		"action": "ci.call.test.v1",
		"by":     "tester",
		"payload": map[string]any{
			"block": true,
		},
	}
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/eval", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.Router().ServeHTTP(rr, req)

	// Expect 403 Forbidden
	require.Equal(t, http.StatusForbidden, rr.Code)

	// Confirm receipt persisted with policy_ref
	rl, err := db.ListReceipts(context.Background(), pool, 10, 0)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rl), 1)

	found := false
	for _, r := range rl {
		if r.Actor == "tester" && r.Type == "ci.call.v1" {
			if pr, ok := r.Payload["policy_ref"].(string); ok && pr != "" {
				found = true
				break
			}
		}
	}
	require.True(t, found, "expected persisted receipt with policy_ref")
}

// Test that allowed action returns 200 and persists policy_ref (when present)
func TestEvalRoute_AllowPath_PersistsPolicyRef(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	gm, err := policy.LoadRegoBundle("../policy/gates.rego")
	require.NoError(t, err)
	rm, err := policy.LoadRegoBundle("../policy/risk.rego")
	require.NoError(t, err)
	fm, err := policy.LoadRegoBundle("../policy/freeze.rego")
	require.NoError(t, err)
	modules := map[string]string{
		"gates.rego":  gm,
		"risk.rego":   rm,
		"freeze.rego": fm,
	}
	eng, err := policy.NewEngine(modules)
	require.NoError(t, err)

	led, err := ledger.Open(context.Background(), "", pool, nil)
	require.NoError(t, err)

	s := NewWithPolicy(pool, led, eng, nil)
	require.NotNil(t, s)
	s.RegisterEvalRoute(eng)

	body := map[string]any{
		"action": "ci.call.test.v1",
		"by":     "tester",
		"payload": map[string]any{
			"block": false,
		},
	}
	buf, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/eval", bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.Router().ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	// Confirm receipt persisted
	rl, err := db.ListReceipts(context.Background(), pool, 10, 0)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rl), 1)
}
