package ledger_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"dis-core/internal/ledger"
)

// TestNewReceipt_CreateSaveVerify exercises the basic lifecycle of a Receipt:
// create → save to disk → read back → verify digital signature.
func TestNewReceipt_CreateSaveVerify(t *testing.T) {
	r := ledger.NewReceipt(
		"domain.terra",
		"unit.test",
		"frozen-core-hash-xyz",
		"console.demo",
		"seat.demo",
	)
	if r == nil {
		t.Fatalf("expected non-nil receipt")
	}
	if r.Hash == "" {
		t.Errorf("expected hash, got empty")
	}
	if r.Signature == "" {
		t.Errorf("expected signature, got empty")
	}

	// Save it and check that the file exists.
	if err := r.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}
	path := filepath.Join("receipts", r.ReceiptID+".json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s to exist, got error: %v", path, err)
	}

	// Read back and re-verify
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	ok, err := ledger.VerifyReceiptJSON(data)
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if !ok {
		t.Errorf("expected verification to succeed")
	}

	// Also test VerifyWithEmbeddedPub on same data
	ok2, err := ledger.VerifyWithEmbeddedPub(data)
	if err != nil {
		t.Fatalf("embedded verify error: %v", err)
	}
	if !ok2 {
		t.Errorf("expected embedded pub verify to succeed")
	}

	// Clean up the receipts dir
	_ = os.RemoveAll("receipts")
}

// TestReceiptJSONRoundTrip ensures JSON marshal/unmarshal integrity.
func TestReceiptJSONRoundTrip(t *testing.T) {
	orig := ledger.NewReceipt("domain.test", "roundtrip", "core-hash", "console.demo", "seat.demo")
	js, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back ledger.Receipt
	if err := json.Unmarshal(js, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.ReceiptID != orig.ReceiptID {
		t.Errorf("ReceiptID mismatch: got %s, want %s", back.ReceiptID, orig.ReceiptID)
	}
}
