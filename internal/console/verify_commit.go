package console

import (
	"dis-core/internal/ledger"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type VerifyReport struct {
	VerifiedAt string           `json:"verified_at"`
	Total      int              `json:"total"`
	Valid      int              `json:"valid"`
	Invalid    int              `json:"invalid"`
	Results    []map[string]any `json:"results"`
}

// CommitVerification creates a signed receipt for an audit run
func (c *Console) CommitVerification(report VerifyReport) (*ledger.Receipt, error) {
	payloadBytes, err := json.Marshal(report)
	if err != nil {
		return nil, err
	}

	// Save a copy of the report itself for traceability
	dir := "versions/v0.6/receipts/audits"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	filename := filepath.Join(dir, fmt.Sprintf("verify_%s.json",
		time.Now().UTC().Format("20060102T150405Z")))
	os.WriteFile(filename, payloadBytes, 0644)

	// Generate the signed receipt
	r := ledger.NewReceipt(c.BoundDomain, "domain.verify.v1", c.BoundCore, c.ID, c.SeatHolders[0])
	if err := ledger.SaveReceipt(r); err != nil {
		return nil, err
	}

	return r, nil
}
