package ledger

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// TestNewReceipt_CreateBasic tests the basic creation of a Receipt.
func TestNewReceipt_CreateBasic(t *testing.T) {
	r, err := NewReceipt(
		"unit.test",            // type
		"domain.terra",         // actor
		"frozen-core-hash-xyz", // target
		"console.demo",         // domain
		"seat.demo",            // payload
	)
	if err != nil {
		t.Fatalf("failed to create receipt: %v", err)
	}
	if r == nil {
		t.Fatalf("expected non-nil receipt")
	}
	if r.ID == "" {
		t.Errorf("expected ID, got empty")
	}
	if r.Type != "unit.test" {
		t.Errorf("expected Type=unit.test, got %s", r.Type)
	}
	if r.Actor != "domain.terra" {
		t.Errorf("expected Actor=domain.terra, got %s", r.Actor)
	}
}

// TestReceiptJSONRoundTrip ensures JSON marshal/unmarshal integrity.
func TestReceiptJSONRoundTrip(t *testing.T) {
	orig, err := NewReceipt("roundtrip", "domain.test", "core-hash", "console.demo", "seat.demo")
	if err != nil {
		t.Fatalf("failed to create receipt: %v", err)
	}
	js, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back Receipt
	if err := json.Unmarshal(js, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.ID != orig.ID {
		t.Errorf("ID mismatch: got %s, want %s", back.ID, orig.ID)
	}
	if back.Type != orig.Type {
		t.Errorf("Type mismatch: got %s, want %s", back.Type, orig.Type)
	}
}

// TestRecordCall_CiCallV1 tests that ci.call.v1 records are properly created
func TestRecordCall_CiCallV1(t *testing.T) {
	// Test the payload structure without database connection
	actor := "domain.test"
	target := "target.test"
	domain := "test"
	action := "test.action.v1"
	payload := map[string]interface{}{
		"test_field": "test_value",
		"count":      42,
	}

	// Create a mock ledger with nil DB to test payload structure
	l := &Ledger{DB: nil}

	// Test that RecordCall returns appropriate error with nil DB
	ctx := context.Background()
	err := l.RecordCall(ctx, actor, target, domain, action, payload)
	if err == nil {
		t.Error("Expected error with nil DB, got nil")
	}
	if err.Error() != "ledger database not initialized" {
		t.Errorf("Expected 'ledger database not initialized' error, got: %v", err)
	}

	// Test CICallPayload structure
	ciPayload := CICallPayload{
		Action:    action,
		Payload:   payload,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}

	// Verify payload structure
	if ciPayload.Action != action {
		t.Errorf("Expected action '%s', got '%s'", action, ciPayload.Action)
	}

	if ciPayload.Payload["test_field"] != "test_value" {
		t.Errorf("Expected test_field 'test_value', got '%v'", ciPayload.Payload["test_field"])
	}

	if ciPayload.Payload["count"] != 42 {
		t.Errorf("Expected count 42, got '%v'", ciPayload.Payload["count"])
	}

	if ciPayload.Timestamp == "" {
		t.Error("Expected non-empty timestamp")
	}

	// Test JSON marshaling
	payloadBytes, err := json.Marshal(ciPayload)
	if err != nil {
		t.Fatalf("Failed to marshal ci.call.v1 payload: %v", err)
	}

	// Test JSON unmarshaling
	var decoded CICallPayload
	err = json.Unmarshal(payloadBytes, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal ci.call.v1 payload: %v", err)
	}

	if decoded.Action != action {
		t.Errorf("Decoded action mismatch: expected '%s', got '%s'", action, decoded.Action)
	}
}

// TestWithProvenance tests context provenance handling
func TestWithProvenance(t *testing.T) {
	ctx := context.Background()

	provenance := []ProvenanceEntry{
		{Type: "domain", Ref: "domain.test", Status: "valid"},
		{Type: "seat", Ref: "seat.test.admin", Status: "valid"},
	}

	// Add provenance to context
	ctxWithProv := WithProvenance(ctx, provenance)

	// Extract provenance from context
	extracted := extractProvenance(ctxWithProv)
	if extracted == nil {
		t.Fatal("Expected provenance, got nil")
	}

	if len(extracted) != 2 {
		t.Fatalf("Expected 2 provenance entries, got %d", len(extracted))
	}

	if extracted[0].Type != "domain" || extracted[0].Ref != "domain.test" {
		t.Errorf("Unexpected first provenance entry: %+v", extracted[0])
	}

	if extracted[1].Type != "seat" || extracted[1].Ref != "seat.test.admin" {
		t.Errorf("Unexpected second provenance entry: %+v", extracted[1])
	}
}
