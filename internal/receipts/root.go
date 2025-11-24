package receipts

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RootSovereignEstablished describes the payload for a root.sovereign.established.v1 receipt.
type RootSovereignEstablished struct {
	RootDomainID      string    `json:"root_domain_id"`
	CorporealDomainID string    `json:"corporeal_domain_id"`
	PactorID          string    `json:"pactor_id"`
	IssuedBy          string    `json:"issued_by"`
	CreatedAt         time.Time `json:"created_at"`
}

// EmitRootSovereignEstablished writes a receipt record inside the provided tx
// indicating that a root sovereign assignment was established.
func EmitRootSovereignEstablished(ctx context.Context, tx pgx.Tx, r RootSovereignEstablished) error {
	if tx == nil {
		return fmt.Errorf("emit root sovereign: tx is nil")
	}

	// canonical payload
	payload := map[string]any{
		"root_domain_id":      r.RootDomainID,
		"corporeal_domain_id": r.CorporealDomainID,
		"pactor_id":           r.PactorID,
		"issued_by":           r.IssuedBy,
		"created_at":          r.CreatedAt,
	}

	// detect receipts schema variant
	var hasTypeCol bool
	if err := tx.QueryRow(ctx, `
        SELECT EXISTS(
            SELECT 1 FROM information_schema.columns
            WHERE table_name = 'receipts' AND column_name = 'type'
        )`).Scan(&hasTypeCol); err != nil {
		return fmt.Errorf("emit root sovereign: inspect receipts schema: %w", err)
	}

	id := uuid.NewString()

	if hasTypeCol {
		if _, err := tx.Exec(ctx, `
            INSERT INTO receipts (id, type, actor, target, domain, payload, created_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
        `, id, "root.sovereign.established.v1", r.IssuedBy, r.RootDomainID, r.RootDomainID, payload, r.CreatedAt); err != nil {
			return fmt.Errorf("emit root sovereign: insert (type schema): %w", err)
		}
		return nil
	}

	if _, err := tx.Exec(ctx, `
        INSERT INTO receipts (id, receipt_type, issued_by, event_id, policy_ref, metadata, issued_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, id, "root.sovereign.established.v1", r.IssuedBy, r.RootDomainID, nil, payload, r.CreatedAt); err != nil {
		return fmt.Errorf("emit root sovereign: insert (alt schema): %w", err)
	}
	return nil
}
