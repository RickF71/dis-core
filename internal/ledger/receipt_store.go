package ledger

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ledgerLock sync.Mutex

// Receipt represents an authoritative, signed event within DIS.
// It records consent lineage, trust feedback, and final moral status.
type Receipt struct {
	ID              string    `json:"id"`
	Action          string    `json:"action"`
	By              string    `json:"by"`
	ConsentRef      string    `json:"consent_ref,omitempty"`
	FeedbackRef     string    `json:"feedback_ref,omitempty"`
	TrustScoreAfter float64   `json:"trust_score_after,omitempty"`
	Status          string    `json:"status"`
	Timestamp       time.Time `json:"timestamp"`
	Hash            string    `json:"hash"`
	Comments        string    `json:"comments,omitempty"`
}

// SaveReceipt marshals the receipt into JSON, writes an individual file,
// and appends it to a rolling ledger.jsonl file.
// It automatically hashes and timestamps each record.
func SaveReceipt(r *Receipt) error {
	ledgerLock.Lock()
	defer ledgerLock.Unlock()

	dir := "receipts"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Fill defaults
	if r.ID == "" {
		r.ID = GenerateUUID()
	}
	if r.Timestamp.IsZero() {
		r.Timestamp = time.Now().UTC()
	}
	if r.Status == "" {
		r.Status = "accepted"
	}

	// Compute hash for integrity
	dataForHash, _ := json.Marshal(r)
	hash := sha256.Sum256(dataForHash)
	r.Hash = fmt.Sprintf("%x", hash[:])

	// Serialize full receipt
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	// --- 1️⃣ Save individual file ---
	filename := filepath.Join(dir, fmt.Sprintf("%s.json", r.ID))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return err
	}

	// --- 2️⃣ Append to rolling ledger file ---
	ledgerFile := filepath.Join(dir, "ledger.jsonl")
	lf, err := os.OpenFile(ledgerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer lf.Close()

	flatData, _ := json.Marshal(r)
	if _, err := lf.Write(append(flatData, '\n')); err != nil {
		return err
	}

	log.Printf("📜 Saved receipt → %s", filename)
	return nil
}

// SaveRawReceipt preserves backward compatibility for any legacy JSON
// payloads that are pre-marshaled.
func SaveRawReceipt(data []byte) error {
	var r Receipt
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	return SaveReceipt(&r)
}
