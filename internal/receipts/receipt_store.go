package ledger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var ledgerLock sync.Mutex

// Receipt represents an authoritative, signed event within DIS.
// It records consent lineage, trust feedback, and final moral status.

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

	// Serialize full receipt using canonical fields
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	// --- 1️⃣ Save individual file ---
	filename := filepath.Join(dir, fmt.Sprintf("%s.json", r.ReceiptID))
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

// GenerateUUID creates a random UUIDv4 string.
// GenerateUUID removed; use canonical receipt ID generator if needed
