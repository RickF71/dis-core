package ledger

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// TestCICallV1_Integration tests the full ci.call.v1 pipeline with actual database
func TestCICallV1_Integration(t *testing.T) {
	// Skip if no test database is available
	ctx := context.Background()
	dsn := "postgres://dis_user:card567@localhost:5432/dis_core_test?sslmode=disable"

	ledger, err := Open(ctx, dsn, nil, nil)
	if err != nil {
		t.Skipf("Skipping integration test - cannot connect to test database: %v", err)
	}
	defer ledger.Close()

	// Test data
	actor := "domain.test.integration"
	target := "target.test.integration"
	domain := "integration"
	action := "test.integration.v1"
	payload := map[string]interface{}{
		"test_type": "integration",
		"operation": "ci_call_pipeline",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"metadata":  map[string]interface{}{"source": "test"},
	}

	// Add provenance to context
	provenance := []ProvenanceEntry{
		{Type: "domain", Ref: "domain.test.integration", Status: "valid"},
		{Type: "test", Ref: "integration.test", Status: "active"},
	}
	ctxWithProvenance := WithProvenance(ctx, provenance)

	// Record the call using ci.call.v1 pipeline
	err = ledger.RecordCall(ctxWithProvenance, actor, target, domain, action, payload)
	if err != nil {
		t.Fatalf("RecordCall failed: %v", err)
	}

	// Query the database to verify the record was inserted correctly
	var id, recordType, recordActor, recordTarget, recordDomain string
	var recordPayload []byte
	var createdAt time.Time

	err = ledger.DB.QueryRow(ctx, `
		SELECT id, type, actor, target, domain, payload, created_at
		FROM receipts
		WHERE type = 'ci.call.v1' AND actor = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, actor).Scan(&id, &recordType, &recordActor, &recordTarget, &recordDomain, &recordPayload, &createdAt)

	if err != nil {
		t.Fatalf("Failed to query inserted record: %v", err)
	}

	// Verify field values
	if recordType != "ci.call.v1" {
		t.Errorf("Expected type 'ci.call.v1', got '%s'", recordType)
	}
	if recordActor != actor {
		t.Errorf("Expected actor '%s', got '%s'", actor, recordActor)
	}
	if recordTarget != target {
		t.Errorf("Expected target '%s', got '%s'", target, recordTarget)
	}
	if recordDomain != domain {
		t.Errorf("Expected domain '%s', got '%s'", domain, recordDomain)
	}

	// Verify ci.call.v1 payload structure
	var ciPayload CICallPayload
	err = json.Unmarshal(recordPayload, &ciPayload)
	if err != nil {
		t.Fatalf("Failed to unmarshal ci.call.v1 payload: %v", err)
	}

	if ciPayload.Action != action {
		t.Errorf("Expected action '%s', got '%s'", action, ciPayload.Action)
	}

	if ciPayload.Payload["test_type"] != "integration" {
		t.Errorf("Expected test_type 'integration', got '%v'", ciPayload.Payload["test_type"])
	}

	if ciPayload.Payload["operation"] != "ci_call_pipeline" {
		t.Errorf("Expected operation 'ci_call_pipeline', got '%v'", ciPayload.Payload["operation"])
	}

	if ciPayload.Timestamp == "" {
		t.Error("Expected non-empty timestamp")
	}

	// Verify provenance was captured
	if len(ciPayload.Provenance) != 2 {
		t.Fatalf("Expected 2 provenance entries, got %d", len(ciPayload.Provenance))
	}

	if ciPayload.Provenance[0].Type != "domain" || ciPayload.Provenance[0].Ref != "domain.test.integration" {
		t.Errorf("Unexpected first provenance entry: %+v", ciPayload.Provenance[0])
	}

	if ciPayload.Provenance[1].Type != "test" || ciPayload.Provenance[1].Ref != "integration.test" {
		t.Errorf("Unexpected second provenance entry: %+v", ciPayload.Provenance[1])
	}

	// Verify the receipt was created recently
	if time.Since(createdAt) > 5*time.Second {
		t.Errorf("Receipt timestamp seems too old: %v", createdAt)
	}

	t.Logf("✅ ci.call.v1 integration test successful - Receipt ID: %s", id)
}

// TestCICallV1_MultipleActions tests multiple actions create separate receipts
func TestCICallV1_MultipleActions(t *testing.T) {
	ctx := context.Background()
	dsn := "postgres://dis_user:card567@localhost:5432/dis_core_test?sslmode=disable"

	ledger, err := Open(ctx, dsn, nil, nil)
	if err != nil {
		t.Skipf("Skipping test - cannot connect to test database: %v", err)
	}
	defer ledger.Close()

	// Record multiple actions
	actions := []struct {
		actor  string
		action string
		data   map[string]interface{}
	}{
		{"domain.test.multi1", "action.create.v1", map[string]interface{}{"type": "create", "count": 1}},
		{"domain.test.multi2", "action.update.v1", map[string]interface{}{"type": "update", "count": 2}},
		{"domain.test.multi3", "action.delete.v1", map[string]interface{}{"type": "delete", "count": 3}},
	}

	for _, action := range actions {
		err = ledger.RecordCall(ctx, action.actor, "target.multi", "multi", action.action, action.data)
		if err != nil {
			t.Fatalf("RecordCall failed for %s: %v", action.actor, err)
		}
	}

	// Query to verify all records were created
	rows, err := ledger.DB.Query(ctx, `
		SELECT actor, payload
		FROM receipts
		WHERE type = 'ci.call.v1' AND target = 'target.multi'
		ORDER BY created_at ASC
	`)
	if err != nil {
		t.Fatalf("Failed to query multiple records: %v", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var actor string
		var payloadBytes []byte
		err = rows.Scan(&actor, &payloadBytes)
		if err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}

		var ciPayload CICallPayload
		err = json.Unmarshal(payloadBytes, &ciPayload)
		if err != nil {
			t.Fatalf("Failed to unmarshal payload: %v", err)
		}

		count++
		expectedActor := actions[count-1].actor
		if actor != expectedActor {
			t.Errorf("Expected actor '%s', got '%s'", expectedActor, actor)
		}

		expectedType := actions[count-1].data["type"]
		if ciPayload.Payload["type"] != expectedType {
			t.Errorf("Expected type '%v', got '%v'", expectedType, ciPayload.Payload["type"])
		}
	}

	if count != 3 {
		t.Errorf("Expected 3 records, got %d", count)
	}

	t.Logf("✅ Multiple ci.call.v1 actions test successful - %d records created", count)
}
