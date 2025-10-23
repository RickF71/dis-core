package ledger

import (
	"database/sql"
	"fmt"
	"time"
)

// EnsureImportReceiptsSchema ensures the import_receipts table exists with a created_at column.
func EnsureImportReceiptsSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS import_receipts (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			target TEXT,
			summary TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	return err
}

type ImportReceipt struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Target    string    `json:"target"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"created_at"`
}

// RecordImport creates a new import receipt.
func (l *Ledger) RecordImport(target, summary string) (*ImportReceipt, error) {
	id := fmt.Sprintf("rcpt-%d", time.Now().UnixNano())
	rec := &ImportReceipt{
		ID:        id,
		Type:      "import.yaml.v1",
		Target:    target,
		Summary:   summary,
		CreatedAt: time.Now(),
	}

	_, err := l.DB.Exec(`
	       INSERT INTO import_receipts (id, type, target, summary, created_at)
	       VALUES ($1, $2, $3, $4, $5)
       `, rec.ID, rec.Type, rec.Target, rec.Summary, rec.CreatedAt)

	return rec, err
}

// ListImports returns the most recent import receipts.
func (l *Ledger) ListImports(limit int) ([]ImportReceipt, error) {
	rows, err := l.DB.Query(`
	       SELECT id, type, target, summary, created_at
	       FROM import_receipts
	       ORDER BY created_at DESC
	       LIMIT $1
       `, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ImportReceipt
	for rows.Next() {
		var r ImportReceipt
		if err := rows.Scan(&r.ID, &r.Type, &r.Target, &r.Summary, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}
