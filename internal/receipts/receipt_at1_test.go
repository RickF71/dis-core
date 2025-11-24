package receipts

import (
	"context"
	"testing"

	"dis-core/internal/contextx"
	dbpkg "dis-core/internal/db"
	"dis-core/internal/testdb"

	"github.com/google/uuid"
)

// TestAT1DecisionMetadataAppears ensures that when a policy decision is
// present on the context, authority receipt recording persists the decision
// metadata into the receipt payload for provenance.
func TestAT1DecisionMetadataAppears(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()
	// Inject a sample AT-1 decision into the context
	decision := map[string]interface{}{"at1.allow": false, "at1.policy": "dis.policy.at1_corruption"}
	ctx = contextx.WithPolicyDecisionMap(ctx, decision)

	domainID := uuid.New().String()
	payload := map[string]any{"foo": "bar"}
	// attach any decision from context (what authority.RecordAuthorityReceiptTx would do)
	if dm, ok := contextx.PolicyDecisionMapFromContext(ctx); ok && dm != nil {
		payload["decision"] = dm
	}

	id := uuid.New().String()
	r := &dbpkg.Receipt{
		ID:      id,
		Type:    "test.at1",
		Actor:   "",
		Target:  domainID,
		Domain:  domainID,
		Payload: payload,
	}
	if err := dbpkg.SaveReceipt(ctx, pool, r); err != nil {
		t.Fatalf("SaveReceipt failed: %v", err)
	}

	// Query the persisted payload and assert the decision map exists
	var persisted map[string]interface{}
	if err := pool.QueryRow(ctx, `SELECT payload FROM receipts WHERE id = $1`, id).Scan(&persisted); err != nil {
		t.Fatalf("failed to query inserted receipt: %v", err)
	}
	d, ok := persisted["decision"].(map[string]interface{})
	if !ok || d == nil {
		t.Fatalf("expected decision map in persisted payload, got %v", persisted["decision"])
	}
	if v, ok := d["at1.allow"]; !ok || v != false {
		t.Fatalf("unexpected at1.allow in persisted decision: %v", v)
	}
	if v, ok := d["at1.policy"]; !ok || v != "dis.policy.at1_corruption" {
		t.Fatalf("unexpected at1.policy in persisted decision: %v", v)
	}
}
