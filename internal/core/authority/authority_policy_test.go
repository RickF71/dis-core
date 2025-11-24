package authority

import (
	"context"
	"testing"

	"dis-core/internal/testdb"

	"github.com/google/uuid"
)

// TestAT1Enforcement ensures EvaluateTx respects the EvalFn decision and
// returns a denial when the policy engine reports deny.
func TestAT1Enforcement(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()
	engine := NewEngine(nil, pool)

	// Stub EvalFn: deny when action == "test.deny"
	engine.SetPolicyEvalFunc(func(ctx context.Context, input map[string]interface{}) (bool, string, map[string]interface{}, error) {
		act, _ := input["action"].(string)
		if act == "test.deny" {
			return false, "at1.deny", map[string]interface{}{"at1.allow": false, "at1.policy": "dis.policy.at1_corruption"}, nil
		}
		return true, "at1.allow", map[string]interface{}{"at1.allow": true, "at1.policy": "dis.policy.at1_corruption"}, nil
	})

	// Begin a tx to pass into EvaluateTx
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	domainID := uuid.New().String()

	allowed, _, details, err := engine.EvaluateTx(ctx, tx, "actor", domainID, "test.deny", map[string]interface{}{})
	if err == nil {
		t.Fatalf("expected EvaluateTx to return error on deny, got nil")
	}
	if allowed {
		t.Fatalf("expected allowed=false on deny")
	}
	if details == nil {
		t.Fatalf("expected details to be returned on deny")
	}
	if v, ok := details["at1.allow"]; !ok || v != false {
		t.Fatalf("expected details.at1.allow == false, got %v", v)
	}
}
