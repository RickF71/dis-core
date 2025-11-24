package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// SaveReceiptTx inserts a receipt using the provided transaction. It detects
// whether the expanded panel columns exist and will include them when present.
func SaveReceiptTx(ctx context.Context, tx pgx.Tx, r *Receipt) error {
	if tx == nil {
		return fmt.Errorf("SaveReceiptTx: tx is nil")
	}

	if r.CreatedAt.IsZero() {
		r.CreatedAt = time.Now().UTC()
	}

	// Detect whether panel columns exist
	var hasPanels bool
	if err := tx.QueryRow(ctx, `
        SELECT EXISTS(
            SELECT 1 FROM information_schema.columns
            WHERE table_name = 'receipts' AND column_name = 'action_panel'
        )`).Scan(&hasPanels); err != nil {
		// On error, fall back to legacy insert
		hasPanels = false
	}

	if hasPanels {
		if r.OriginDomainID == "" {
			return fmt.Errorf("receipt missing origin domain")
		}
		_, err := tx.Exec(ctx, `
            INSERT INTO receipts (id, type, actor, target, domain, payload, created_at, origin_domain_id, origin_domain_name,
                action_panel, policy_panel, identity_panel, dimension_panel, lineage_panel, domain_panel)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
        `, r.ID, r.Type, r.Actor, r.Target, r.Domain, r.Payload, r.CreatedAt, r.OriginDomainID, r.OriginDomainName,
			r.ActionPanel, r.PolicyPanel, r.IdentityPanel, r.DimensionPanel, r.LineagePanel, r.DomainPanel)
		if err != nil {
			return fmt.Errorf("SaveReceiptTx: failed to insert receipt with panels: %w", err)
		}
		return nil
	}

	// Legacy insert (no panel/origin columns)
	_, err := tx.Exec(ctx, `
        INSERT INTO receipts (id, type, actor, target, domain, payload, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, r.ID, r.Type, r.Actor, r.Target, r.Domain, r.Payload, r.CreatedAt)
	if err != nil {
		return fmt.Errorf("SaveReceiptTx: failed legacy insert: %w", err)
	}
	return nil
}
