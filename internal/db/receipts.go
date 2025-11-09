package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Receipt matches the current receipts table schema.
type Receipt struct {
	ID        string         `json:"id"`   // UUID / text primary key
	Type      string         `json:"type"` // e.g. "ci.call.v1"
	Actor     string         `json:"actor"`
	Target    string         `json:"target"`
	Domain    string         `json:"domain"`
	Payload   map[string]any `json:"payload"` // JSONB in DB
	CreatedAt time.Time      `json:"created_at"`
}

// EnsureReceiptsSchema creates the receipts table if missing (safety net for dev/test).
func EnsureReceiptsSchema(ctx context.Context, pool *pgxpool.Pool) error {
	schema := `
	CREATE TABLE IF NOT EXISTS receipts (
		id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
		type TEXT NOT NULL,
		actor TEXT,
		target TEXT,
		domain TEXT,
		payload JSONB,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	CREATE INDEX IF NOT EXISTS idx_receipts_created_at ON receipts(created_at);
	`
	_, err := pool.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to ensure receipts table: %w", err)
	}
	fmt.Println("✅ receipts table verified or created (Postgres).")
	return nil
}

// SaveReceipt inserts a new receipt into the receipts table.
func SaveReceipt(ctx context.Context, pool *pgxpool.Pool, r *Receipt) error {
	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO receipts (id, type, actor, target, domain, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`, r.ID, r.Type, r.Actor, r.Target, r.Domain, r.Payload, r.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert receipt: %w", err)
	}
	return nil
}

// ListReceipts fetches recent receipts with optional limit and offset.
func ListReceipts(ctx context.Context, pool *pgxpool.Pool, limit int, offset int) ([]Receipt, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := pool.Query(ctx, `
		SELECT id, type, actor, target, domain, payload, created_at
		FROM receipts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2;
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list receipts: %w", err)
	}
	defer rows.Close()

	var out []Receipt
	for rows.Next() {
		var r Receipt
		if err := rows.Scan(&r.ID, &r.Type, &r.Actor, &r.Target, &r.Domain, &r.Payload, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// CountReceipts returns total count of receipts in the database.
func CountReceipts(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	var n int64
	err := pool.QueryRow(ctx, `SELECT COUNT(1) FROM receipts;`).Scan(&n)
	return n, err
}
