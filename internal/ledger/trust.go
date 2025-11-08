package ledger

import (
	"database/sql"
	"time"
)

// TrustEntry represents one verification event between peers.
type TrustEntry struct {
	Peer       string
	Action     string // "sent" or "received"
	Status     string // "ok", "fail", "unreachable"
	ReceiptID  string
	CoreHash   string
	VerifiedAt time.Time
	Notes      string
}

// TrustLedger provides DB-backed persistence for trust events.
type TrustLedger struct {
	DB *sql.DB
}

// OpenTrustLedger ensures the trust_entries table exists and returns a handle.
func OpenTrustLedger(db *sql.DB) (*TrustLedger, error) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS trust_entries (
			id SERIAL PRIMARY KEY,
			peer TEXT,
			action TEXT,
			status TEXT,
			receipt_id TEXT,
			core_hash TEXT,
			verified_at TIMESTAMPTZ DEFAULT now(),
			notes TEXT
		);
	`)
	if err != nil {
		return nil, err
	}
	return &TrustLedger{DB: db}, nil
}

// Add inserts a new trust entry.
func (l *TrustLedger) Add(entry TrustEntry) error {
	_, err := l.DB.Exec(`
		INSERT INTO trust_entries (peer, action, status, receipt_id, core_hash, verified_at, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, entry.Peer, entry.Action, entry.Status, entry.ReceiptID, entry.CoreHash, entry.VerifiedAt, entry.Notes)
	return err
}

// List returns the most recent trust entries.
func (l *TrustLedger) List(limit int) ([]TrustEntry, error) {
	rows, err := l.DB.Query(`
		SELECT peer, action, status, receipt_id, core_hash, verified_at, notes
		FROM trust_entries
		ORDER BY verified_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TrustEntry
	for rows.Next() {
		var e TrustEntry
		if err := rows.Scan(&e.Peer, &e.Action, &e.Status, &e.ReceiptID, &e.CoreHash, &e.VerifiedAt, &e.Notes); err != nil {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}
