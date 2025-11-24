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
	// Optional origin fields
	OriginDomainID   string `json:"origin_domain_id"`
	OriginDomainName string `json:"origin_domain_name"`
	// Panel columns for ReceiptEnvelope
	ActionPanel    map[string]any `json:"action_panel"`
	PolicyPanel    map[string]any `json:"policy_panel"`
	IdentityPanel  map[string]any `json:"identity_panel"`
	DimensionPanel map[string]any `json:"dimension_panel"`
	LineagePanel   map[string]any `json:"lineage_panel"`
	DomainPanel    map[string]any `json:"domain_panel"`
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
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		origin_domain_id TEXT,
		origin_domain_name TEXT,
		action_panel JSONB,
		policy_panel JSONB,
		identity_panel JSONB,
		dimension_panel JSONB,
		lineage_panel JSONB,
		domain_panel JSONB
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
	if pool == nil {
		return fmt.Errorf("SaveReceipt: db pool is nil")
	}

	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("SaveReceipt: begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := SaveReceiptTx(ctx, tx, r); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("SaveReceipt: commit tx: %w", err)
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
		var domainPtr *string
		if err := rows.Scan(&r.ID, &r.Type, &r.Actor, &r.Target, &domainPtr, &r.Payload, &r.CreatedAt); err != nil {
			return nil, err
		}
		if domainPtr != nil {
			r.Domain = *domainPtr
		} else {
			r.Domain = ""
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
