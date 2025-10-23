package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Receipt matches the current receipts table schema.
type Receipt struct {
	ID        int64     `json:"id"`
	ReceiptID string    `json:"receipt_id"`
	SchemaRef string    `json:"schema_ref"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// EnsureReceiptsSchema creates the receipts table if missing (for safety).
func EnsureReceiptsSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS receipts (
		id SERIAL PRIMARY KEY,
		receipt_id TEXT UNIQUE NOT NULL,
		schema_ref TEXT,
		content TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	CREATE INDEX IF NOT EXISTS idx_receipts_created_at ON receipts(created_at);
	`
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to ensure receipts table: %w", err)
	}
	fmt.Println("✅ receipts table verified or created (Postgres).")
	return nil
}

// InsertReceipt adds a new receipt entry into the receipts table.
func InsertReceipt(db *sql.DB, r *Receipt) (int64, error) {
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	q := `
       INSERT INTO receipts (receipt_id, schema_ref, content, created_at)
       VALUES ($1, $2, $3, $4)
       RETURNING id;
       `
	var id int64
	err := db.QueryRow(q, r.ReceiptID, r.SchemaRef, r.Content, r.CreatedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert receipt: %w", err)
	}
	return id, nil
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

	rows, err := db.QueryContext(context.Background(), `
	       SELECT id, receipt_id, schema_ref, content, created_at
	       FROM receipts
	       ORDER BY id DESC
	       LIMIT $1 OFFSET $2;
       `, opts.Limit, opts.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list receipts: %w", err)
	}
	defer rows.Close()

	var out []Receipt
	for rows.Next() {
		var r Receipt
		if err := rows.Scan(&r.ID, &r.ReceiptID, &r.SchemaRef, &r.Content, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

//
// === v0.9.3 Additions for AutoRevocation + /api/status ===
//

// SaveReceipt inserts a receipt directly using the default DB handle.
func SaveReceipt(r Receipt) error {
	if DefaultDB == nil {
		return fmt.Errorf("db not initialized")
	}
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	_, err := DefaultDB.Exec(`
	       INSERT INTO receipts (receipt_id, schema_ref, content, created_at)
	       VALUES ($1, $2, $3, $4);
       `, r.ReceiptID, r.SchemaRef, r.Content, r.CreatedAt)
	return err
}

// CountReceipts returns total count of receipts in the database.
func CountReceipts() (int64, error) {
	if DefaultDB == nil {
		return 0, fmt.Errorf("db not initialized")
	}
	var n int64
	err := DefaultDB.QueryRow(`SELECT COUNT(1) FROM receipts;`).Scan(&n)
	return n, err
}
