package main

import (
	"dis-core/internal/db"
	"dis-core/internal/ledger"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type AuditResult struct {
	File       string `json:"file"`
	Status     string `json:"status"`
	Reason     string `json:"reason"`
	VerifiedAt string `json:"verified_at"`
}

func main() {
	dir := "versions/v0.6/receipts/generated"
	archiveDir := "versions/v0.6/receipts/archive"
	quarantineDir := "versions/v0.6/receipts/quarantine"

	os.MkdirAll(archiveDir, 0755)
	os.MkdirAll(quarantineDir, 0755)

	files := []string{}
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
			files = append(files, path)
		}
		return nil
	})

	results := []AuditResult{}
	validCount, invalidCount := 0, 0

	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			results = append(results, AuditResult{
				File: filepath.Base(path), Status: "error",
				Reason: err.Error(), VerifiedAt: db.NowRFC3339Nano(),
			})
			continue
		}

		ok, err := ledger.VerifyWithEmbeddedPub(data)
		if err != nil {
			// Determine if legacy (no signature) or corrupted
			if strings.Contains(err.Error(), "insufficient data") {
				os.Rename(path, filepath.Join(archiveDir, filepath.Base(path)))
				results = append(results, AuditResult{
					File:       filepath.Base(path),
					Status:     "archived",
					Reason:     "legacy / unsigned receipt",
					VerifiedAt: db.NowRFC3339Nano(),
				})
			} else {
				os.Rename(path, filepath.Join(quarantineDir, filepath.Base(path)))
				results = append(results, AuditResult{
					File:       filepath.Base(path),
					Status:     "quarantined",
					Reason:     err.Error(),
					VerifiedAt: db.NowRFC3339Nano(),
				})
				invalidCount++
			}
			continue
		}

		if ok {
			results = append(results, AuditResult{
				File: filepath.Base(path), Status: "valid",
				Reason: "signature verified", VerifiedAt: db.NowRFC3339Nano(),
			})
			validCount++
		} else {
			os.Rename(path, filepath.Join(quarantineDir, filepath.Base(path)))
			results = append(results, AuditResult{
				File: filepath.Base(path), Status: "quarantined",
				Reason: "signature mismatch", VerifiedAt: db.NowRFC3339Nano(),
			})
			invalidCount++
		}
	}

	// Write the ledger
	ledger := map[string]any{
		"verified_at":   db.NowRFC3339Nano(),
		"total":         len(files),
		"valid":         validCount,
		"invalid":       invalidCount,
		"audit_results": results,
	}

	outPath := "versions/v0.6/receipts/verification_ledger.json"
	f, _ := os.Create(outPath)
	defer f.Close()
	json.NewEncoder(f).Encode(ledger)

	fmt.Printf("✨ Verification completed: %d valid, %d invalid\n", validCount, invalidCount)
	fmt.Printf("🪶 Ledger written to %s\n", outPath)
}
