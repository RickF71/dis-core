package console

import (
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dis-core/internal/db"
	"dis-core/internal/receipts"
)

var lastVerificationTime time.Time
var lastVerificationFile = "versions/v0.6/receipts/last_verification.txt"

func (c *Console) RunVerification() (VerifyReport, *receipts.Receipt, error) {
	dir := "versions/v0.6/receipts/generated"

	// Gather receipt files
	files := []string{}
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
			files = append(files, path)
		}
		return nil
	})

	validCount := 0
	invalidCount := 0
	results := []map[string]any{}

	for _, path := range files {
		data, err := os.ReadFile(path)
		status := "valid"
		reason := "signature verified"

		if err != nil {
			status = "error"
			reason = err.Error()
		} else {
			ok, verr := receipts.VerifyWithEmbeddedPub(data)
			if verr != nil || !ok {
				status = "invalid"
				reason = "signature mismatch"
				invalidCount++
			} else {
				validCount++
			}
		}

		results = append(results, map[string]any{
			"file":        filepath.Base(path),
			"status":      status,
			"reason":      reason,
			"verified_at": db.NowRFC3339Nano(),
		})
	}

	report := VerifyReport{
		VerifiedAt: db.NowRFC3339Nano(),
		Total:      len(files),
		Valid:      validCount,
		Invalid:    invalidCount,
		Results:    results,
	}

	// Keep a human-readable copy of the report
	// (CommitVerification also saves a copy and issues a signed receipt)
	b, _ := json.MarshalIndent(report, "", "  ")
	_ = os.MkdirAll("versions/v0.6/receipts/audits", 0755)
	_ = os.WriteFile(
		filepath.Join("versions/v0.6/receipts/audits", "verify_"+time.Now().UTC().Format("20060102T150405Z")+".json"),
		b, 0644,
	)

	// Issue signed verification receipt
	rcpt, err := c.CommitVerification(report)
	if err != nil {
		return report, nil, err
	}
	return report, rcpt, nil
}

func (c *Console) RunVerificationIfNeeded() (bool, VerifyReport, *receipts.Receipt, error) {
	dir := "versions/v0.6/receipts/generated"

	// Get latest modification time among all receipts
	var latestMod time.Time
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, err := os.Stat(path); err == nil {
			if info.ModTime().After(latestMod) {
				latestMod = info.ModTime()
			}
		}
		return nil
	})

	// Add a 5-second buffer to avoid near-simultaneous modtimes triggering a new audit
	buffered := lastVerificationTime.Add(5 * time.Second)
	if !lastVerificationTime.IsZero() && !latestMod.After(buffered) {
		log.Println("🕒 No new receipts detected since last verification; skipping.")
		return false, VerifyReport{}, nil, nil
	}

	// Perform full verification
	report, receipt, err := c.RunVerification()
	if err == nil {
		lastVerificationTime = time.Now().UTC()
		saveLastVerification(lastVerificationTime)
	}
	return true, report, receipt, err
}

// --- Persistence for last verification timestamp ---
// LoadLastVerification reads timestamp from disk at startup.
func LoadLastVerification() {
	data, err := os.ReadFile(lastVerificationFile)
	if err != nil {
		log.Println("ℹ️ No last_verification.txt found; will verify on next cycle.")
		return
	}
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(data)))
	if err != nil {
		log.Printf("⚠️ Failed to parse last_verification.txt: %v", err)
		return
	}
	lastVerificationTime = t
	log.Printf("📂 Loaded last verification timestamp: %s", t.UTC().Format(time.RFC3339))
}

// saveLastVerification writes the verification timestamp to disk.
func saveLastVerification(t time.Time) {
	_ = os.MkdirAll(filepath.Dir(lastVerificationFile), 0755)
	_ = os.WriteFile(lastVerificationFile, []byte(t.UTC().Format(time.RFC3339)), 0644)
}
