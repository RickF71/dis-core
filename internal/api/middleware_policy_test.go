package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"dis-core/internal/api/middleware"
	authpkg "dis-core/internal/auth"
	"dis-core/internal/policy"
	"dis-core/internal/testdb"

	"github.com/google/uuid"
)

// fake policy engine implementing minimal EvaluateAction for tests

// TestPolicyDenyViaMiddleware ensures middleware denies and handler isn't invoked.
func TestPolicyDenyViaMiddleware(t *testing.T) {
	// This test verifies middleware returns 403 for denied decisions. To keep
	// scope focused we simulate the effect by creating a handler wrapped by
	// the enforcement middleware but we inject a small policy engine via the
	// WithPolicyEngine middleware that returns a denied decision.

	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	// simple handler that would create a receipt if called
	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	// Build router with policy engine and enforcement middleware
	// For this focused unit test we reuse the real policy engine implementation
	// by constructing a small in-place engine that always denies.
	denyEngine := &testPolicyEngine{allow: false, reason: "denied-for-test"}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	// attach an active user so middleware can read actor info if needed
	au := &authpkg.ActiveUser{CorporealDomainUID: uuid.NewString(), CorporealDomainID: 1}
	req = req.WithContext(authpkg.WithActiveUser(req.Context(), au))

	// wrap handler
	wrapped := middleware.WithPolicyEngine(denyEngine)(middleware.WithPolicyEnforcement()(h))
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden, got %d", rr.Code)
	}
	if called {
		t.Fatalf("handler should not have been called on denied action")
	}
}

// testPolicyEngine is a tiny policy.PolicyEngine used for middleware tests.
type testPolicyEngine struct {
	allow  bool
	reason string
}

// EvaluateAction implements the policy.PolicyEngine interface used by middleware.
func (tpe *testPolicyEngine) EvaluateAction(ctx context.Context, input map[string]interface{}) (*policy.PolicyDecision, error) {
	pd := &policy.PolicyDecision{
		Allow:  tpe.allow,
		Reason: tpe.reason,
	}
	return pd, nil
}
