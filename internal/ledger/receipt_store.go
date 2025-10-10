package ledger

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ledgerLock sync.Mutex

// Receipt is a minimal placeholder for ledger entries.
// Replace or expand with your existing receipt struct when integrated.
type Receipt struct {
	ID        string
	Action    string
	Timestamp string
	Hash      string
}

// SaveReceipt writes receipt JSON bytes both as an individual JSON file
// and as a single-line entry in a rolling ledger file (ledger.jsonl).
func SaveReceipt(data []byte) error {
	dir := "receipts"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// --- 1️⃣ Save individual file ---
	filename := filepath.Join(dir, time.Now().Format("2006-01-02T15-04-05.000")+".json")

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	f.Close()
	if err != nil {
		return err
	}

	// --- 2️⃣ Append to rolling ledger file ---
	ledgerFile := filepath.Join(dir, "ledger.jsonl")

	ledgerLock.Lock()
	defer ledgerLock.Unlock()

	lf, err := os.OpenFile(ledgerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer lf.Close()

	// Each receipt is stored as a JSON line
	if _, err := lf.Write(append(data, '\n')); err != nil {
		return err
	}

	log.Printf("🧾 Receipt saved: %s", filename)
	return nil
}
