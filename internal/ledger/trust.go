package ledger

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// TrustEntry represents one verification event between peers.
type TrustEntry struct {
	Peer       string    `json:"peer"`
	Action     string    `json:"action"` // "sent" or "received"
	Status     string    `json:"status"` // "ok", "fail", "unreachable"
	ReceiptID  string    `json:"receipt_id"`
	CoreHash   string    `json:"core_hash"`
	VerifiedAt time.Time `json:"verified_at"`
	Notes      string    `json:"notes,omitempty"`
}

// TrustLedger stores all trust events persistently.
type TrustLedger struct {
	mu      sync.Mutex
	Entries []TrustEntry `json:"entries"`
	Path    string       `json:"-"`
}

// LoadTrustLedger loads an existing ledger or creates a new one.
func LoadTrustLedger(path string) (*TrustLedger, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &TrustLedger{Entries: []TrustEntry{}, Path: path}, nil
		}
		return nil, err
	}

	var l TrustLedger
	if err := json.Unmarshal(data, &l); err != nil {
		return nil, err
	}
	l.Path = path
	return &l, nil
}

// Add appends a new entry and writes the updated ledger back to disk.
func (l *TrustLedger) Add(entry TrustEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Entries = append(l.Entries, entry)
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(l.Path, data, 0644)
}
