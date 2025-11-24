package authority

import (
	"context"
	"testing"

	dbstore "dis-core/internal/db"
	"dis-core/internal/testdb"

	"github.com/google/uuid"
)

// Test that EvaluateTx integration prevents a domain freeze when policy denies
func TestEvaluateTx_DeniesFreezeAndRollsBack(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	defer pool.Close()

	ctx := context.Background()

	// ensure there is at least one domain
	var domainID uuid.UUID
	if err := pool.QueryRow(ctx, "SELECT id FROM domains LIMIT 1").Scan(&domainID); err != nil {
		domainID = uuid.New()
		if _, err := pool.Exec(ctx, `INSERT INTO domains (id, name, domain_type, authority, created_at) VALUES ($1,$2,$3,$4,NOW())`, domainID, "eval-test", "corporeal", "test"); err != nil {
			t.Fatalf("failed to create domain: %v", err)
		}
	}

	eng := NewEngine(nil, pool)

	// attach a simple EvalFn that denies freezes with scope == "deny-scope"
	eng.SetPolicyEvalFunc(func(ctx context.Context, input map[string]interface{}) (bool, string, map[string]interface{}, error) {
		if a, ok := input["action"].(string); ok && a == "domain.freeze.v1" {
			if p, ok := input["payload"].(map[string]interface{}); ok {
				if s, sok := p["scope"].(string); sok && s == "deny-scope" {
					details := map[string]interface{}{"policy_ref": "ci_rules:ci_call_test_block_v1", "gates.deny_code": "at1.ci_call_test_block_v1"}
					return false, "deny:at1.ci_call_test_block_v1", details, nil
				}
			}
		}
		return true, "allow", nil, nil
	})

	// attempt to freeze with denied scope
	_, err := eng.FreezeDomain(ctx, domainID, "deny-scope", "reason", uuid.New(), nil)
	if err == nil {
		t.Fatalf("expected FreezeDomain to be denied by policy")
	}

	// ensure no active freeze exists for that scope (use store helper)
	frs, ferr := dbstore.GetActiveFreezes(ctx, pool, domainID.String())
	if ferr != nil {
		t.Fatalf("GetActiveFreezes failed: %v", ferr)
	}
	for _, f := range frs {
		if f.Scope == "deny-scope" {
			t.Fatalf("expected no active freeze row after denied FreezeDomain")
		}
	}
}
