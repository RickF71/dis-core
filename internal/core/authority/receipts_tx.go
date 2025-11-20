package authority

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RecordAuthorityReceiptTx records a receipt inside the provided tx for the given domain.
// This is a tx-aware canonical helper used by orchestrators that already manage transactions.
func RecordAuthorityReceiptTx(ctx context.Context, tx pgx.Tx, domainID string, typ string, payload map[string]any) (string, error) {
	// attach domain id
	if payload == nil {
		payload = map[string]any{}
	}
	payload["domain_id"] = domainID

	// determine previous hash if payload column exists
	var prevHash string
	var hasPayloadCol bool
	if err := tx.QueryRow(ctx, `
        SELECT EXISTS(
            SELECT 1 FROM information_schema.columns
            WHERE table_name = 'receipts' AND column_name = 'payload'
        )`).Scan(&hasPayloadCol); err == nil && hasPayloadCol {
		rows, err := tx.Query(ctx, `SELECT payload FROM receipts WHERE domain = $1 ORDER BY created_at ASC`, domainID)
		if err == nil {
			defer rows.Close()
			var last map[string]any
			for rows.Next() {
				_ = rows.Scan(&last)
			}
			if last != nil {
				if h, ok := last["hash"].(string); ok {
					prevHash = h
				}
			}
		}
	}
	payload["prev_hash"] = prevHash

	// detect schema variant
	var hasTypeCol bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'receipts' AND column_name = 'type'
		)`).Scan(&hasTypeCol); err != nil {
		return "", fmt.Errorf("record authority receipt tx: inspect receipts schema: %w", err)
	}

	id := uuid.NewString()
	now := time.Now().UTC()
	actor := ""

	if hasTypeCol {
		if _, err := tx.Exec(ctx, `
			INSERT INTO receipts (id, type, actor, target, domain, payload, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, id, typ, actor, domainID, domainID, payload, now); err != nil {
			return "", fmt.Errorf("record authority receipt tx: insert (type schema): %w", err)
		}
		return id, nil
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO receipts (id, receipt_type, issued_by, event_id, policy_ref, metadata, issued_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, id, typ, actor, domainID, payload["policy_ref"], payload, now); err != nil {
		return "", fmt.Errorf("record authority receipt tx: insert (alt schema): %w", err)
	}
	return id, nil
}
