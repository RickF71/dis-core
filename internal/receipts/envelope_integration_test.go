package receipts

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This integration test asserts that SaveEnvelope writes panels to the
// ledger file. It intentionally does not require a running Postgres
// instance; DB-level assertions can be enabled via DATABASE_URL.
func TestSaveEnvelope_WritesLedger(t *testing.T) {
	// Create a temporary receipt directory to avoid clobbering project files.
	orig := "receipts"
	tmpDir := filepath.Join(os.TempDir(), "dis_test_receipts")
	os.RemoveAll(tmpDir)
	os.MkdirAll(tmpDir, 0755)
	// change working dir for the duration
	prevWd, _ := os.Getwd()
	defer os.Chdir(prevWd)
	os.Chdir(tmpDir)

	originID := "dom-int"
	originName := "Integration"
	env := NewEnvelope(originID, originName, "actor-int")
	env.ActionPanel["type"] = "test.action"
	env.PolicyPanel["at1.allow"] = true
	env.Timestamp = time.Now().UTC()

	if err := SaveEnvelope(env); err != nil {
		t.Fatalf("SaveEnvelope failed: %v", err)
	}

	ledgerFile := filepath.Join(tmpDir, "receipts", "ledger.jsonl")
	f, err := os.Open(ledgerFile)
	if err != nil {
		t.Fatalf("failed to open ledger: %v", err)
	}
	defer f.Close()
	b, _ := io.ReadAll(f)
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) == 0 {
		t.Fatalf("ledger is empty")
	}
	var last map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("failed to parse ledger last line: %v", err)
	}
	if _, ok := last["action"]; !ok {
		t.Fatalf("expected action panel in ledger entry")
	}
	if _, ok := last["policy"]; !ok {
		t.Fatalf("expected policy panel in ledger entry")
	}

	// cleanup
	os.RemoveAll(tmpDir)
	_ = orig
}
