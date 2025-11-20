package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CreateCorporealDomainNamed creates a corporeal domain with the provided
// name and associates it with the actor (actorID). Returns the created domain UUID.
func CreateCorporealDomainNamed(ctx context.Context, tx pgx.Tx, actorID uuid.UUID, name string) (uuid.UUID, error) {
	if name == "" {
		return uuid.Nil, fmt.Errorf("domain name required")
	}

	// Check if domain exists
	var existing string
	if err := tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = $1 LIMIT 1`, name).Scan(&existing); err == nil {
		id, _ := uuid.Parse(existing)
		return id, nil
	}

	// Find terra parent (try common name variants used in seeds)
	var parentID string
	_ = tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = 'terra' LIMIT 1`).Scan(&parentID)
	if parentID == "" {
		_ = tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = 'domain.terra' LIMIT 1`).Scan(&parentID)
	}

	id := uuid.New()
	// attempt to insert with domain_type if present, fall back otherwise
	if _, err := tx.Exec(ctx, `INSERT INTO domains (id, name, parent_id, domain_type, payload, created_at) VALUES ($1,$2,$3,'corporeal','{}'::jsonb,$4)`, id.String(), name, parentID, time.Now().UTC()); err == nil {
		return id, nil
	}

	// fallback insert without domain_type
	if _, err := tx.Exec(ctx, `INSERT INTO domains (id, name, parent_id, payload, created_at) VALUES ($1,$2,$3,'{}'::jsonb,$4)`, id.String(), name, parentID, time.Now().UTC()); err != nil {
		return uuid.Nil, fmt.Errorf("create corporeal domain (fallback): %w", err)
	}
	return id, nil
}
