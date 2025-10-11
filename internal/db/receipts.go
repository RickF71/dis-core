package db

import (
	"context"
	"database/sql"
	"time"
)

// Receipt matches the current receipts table schema.
type Receipt struct {
	ID        int64  `json:"id"`
	ReceiptID string `json:"receipt_id"`
	SchemaRef string `json:"schema_ref"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// EnsureReceiptsSchema creates the receipts table if missing (for safety).
func EnsureReceiptsSchema(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS receipts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	receipt_id TEXT UNIQUE NOT NULL,
	schema_ref TEXT,
	content TEXT,
	timestamp TEXT
);
CREATE INDEX IF NOT EXISTS idx_receipts_ts ON receipts(timestamp);
`
	_, err := db.Exec(schema)
	return err
}

// InsertReceipt adds a new receipt entry into the receipts table.
func InsertReceipt(db *sql.DB, r *Receipt) (int64, error) {
	if r.Timestamp == "" {
		r.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	q := `INSERT INTO receipts(receipt_id, schema_ref, content, timestamp)
	      VALUES(?, ?, ?, ?)`
	res, err := db.Exec(q, r.ReceiptID, r.SchemaRef, r.Content, r.Timestamp)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListOpts provides filtering/pagination parameters.
type ListOpts struct {
	Limit  int
	Offset int
}

// ListReceipts fetches recent receipts with optional limit/offset.
func ListReceipts(db *sql.DB, opts ListOpts) ([]Receipt, error) {
	if opts.Limit <= 0 || opts.Limit > 500 {
		opts.Limit = 100
	}
	if opts.Offset < 0 {
		opts.Offset = 0
	}

	rows, err := db.QueryContext(context.Background(),
		`SELECT id, receipt_id, schema_ref, content, timestamp
		 FROM receipts
		 ORDER BY id DESC
		 LIMIT ? OFFSET ?`, opts.Limit, opts.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Receipt
	for rows.Next() {
		var r Receipt
		if err := rows.Scan(&r.ID, &r.ReceiptID, &r.SchemaRef, &r.Content, &r.Timestamp); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
