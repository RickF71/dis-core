package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CreateCorporealDomainFromSubject creates a new corporeal domain row for the
// given subject within the provided transaction. It returns the domain id.
// This is the canonical helper to instantiate corporeal domains (avoids ad-hoc SQL
// scattered across the codebase).
func CreateCorporealDomainFromSubject(ctx context.Context, tx pgx.Tx, subject string) (string, error) {
	domainName := fmt.Sprintf("domain.corporeal.%s", subject)

	// Check if domain already exists
	var existing string
	if err := tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = $1 LIMIT 1`, domainName).Scan(&existing); err == nil {
		return existing, nil
	}

	// Find terra parent id
	var parentID string
	_ = tx.QueryRow(ctx, `SELECT id FROM domains WHERE id = 'domain.terra' LIMIT 1`).Scan(&parentID)
	if parentID == "" {
		_ = tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = 'domain.terra' LIMIT 1`).Scan(&parentID)
	}

	id := uuid.NewString()
	payload := map[string]any{"meta": map[string]any{"origin": "know_thyself.atomic.v1", "subject": subject}}

	if _, err := tx.Exec(ctx, `INSERT INTO domains (id, name, parent_id, payload, created_at) VALUES ($1,$2,$3,$4,$5)`, id, domainName, parentID, payload, time.Now().UTC()); err != nil {
		return "", fmt.Errorf("create corporeal domain: %w", err)
	}
	return id, nil
}
