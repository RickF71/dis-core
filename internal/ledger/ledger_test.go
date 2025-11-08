package ledger

import (
	"encoding/json"
	"testing"
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
