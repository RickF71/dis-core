package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FreezeRecord is a DB-level representation of a domain freeze row
type FreezeRecord struct {
	ID         string     `json:"id"`
	DomainID   string     `json:"domain_id"`
	Scope      string     `json:"scope"`
	Reason     string     `json:"reason"`
	CreatedBy  string     `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
	TTLUntil   *time.Time `json:"ttl_until,omitempty"`
	OverrideOf *string    `json:"override_of,omitempty"`
	IsActive   bool       `json:"is_active"`
}

// InsertDomainFreezeTx inserts a new domain_freeze_state row inside tx.
func InsertDomainFreezeTx(ctx context.Context, tx pgx.Tx, domainID, scope, reason, createdBy string, ttl *time.Time, overrideOf *string) (string, error) {
	id := ""
	err := tx.QueryRow(ctx, `
        INSERT INTO domain_freeze_state (domain_id, scope, reason, created_by, ttl_until, override_of, is_active)
        VALUES ($1::uuid, $2, $3, $4::uuid, $5, $6::uuid, true)
        RETURNING id
    `, domainID, scope, reason, createdBy, ttl, overrideOf).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert domain freeze (tx): %w", err)
	}
	return id, nil
}

// DeactivateDomainFreezeTx marks active freeze rows for domain+scope as inactive
func DeactivateDomainFreezeTx(ctx context.Context, tx pgx.Tx, domainID, scope string) error {
	_, err := tx.Exec(ctx, `
        UPDATE domain_freeze_state
        SET is_active = FALSE
        WHERE domain_id = $1::uuid AND scope = $2 AND is_active = TRUE
    `, domainID, scope)
	if err != nil {
		return fmt.Errorf("deactivate domain freeze (tx): %w", err)
	}
	return nil
}

// InsertDomainFreezeOverrideTx records an override and returns its id.
func InsertDomainFreezeOverrideTx(ctx context.Context, tx pgx.Tx, domainID, scope, reason, createdBy string, overrideOf string) (string, error) {
	id := ""
	err := tx.QueryRow(ctx, `
        INSERT INTO domain_freeze_state (domain_id, scope, reason, created_by, ttl_until, override_of, is_active)
        VALUES ($1::uuid, $2, $3, $4::uuid, NULL, $5::uuid, TRUE)
        RETURNING id
    `, domainID, scope, reason, createdBy, overrideOf).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert domain freeze override (tx): %w", err)
	}
	return id, nil
}

// GetActiveFreezes fetches active freeze records for a domain (ttl filter applied).
func GetActiveFreezes(ctx context.Context, db *pgxpool.Pool, domainID string) ([]FreezeRecord, error) {
	rows, err := db.Query(ctx, `
        SELECT id::text, domain_id::text, scope, reason, created_by::text, created_at, ttl_until, override_of::text, is_active
        FROM domain_freeze_state
        WHERE domain_id = $1::uuid AND is_active = TRUE AND (ttl_until IS NULL OR ttl_until > now())
        ORDER BY created_at DESC
    `, domainID)
	if err != nil {
		return nil, fmt.Errorf("get active freezes: %w", err)
	}
	defer rows.Close()

	var out []FreezeRecord
	for rows.Next() {
		var fr FreezeRecord
		var overrideOf sql.NullString
		if err := rows.Scan(&fr.ID, &fr.DomainID, &fr.Scope, &fr.Reason, &fr.CreatedBy, &fr.CreatedAt, &fr.TTLUntil, &overrideOf, &fr.IsActive); err != nil {
			return nil, err
		}
		if overrideOf.Valid {
			v := overrideOf.String
			fr.OverrideOf = &v
		}
		out = append(out, fr)
	}
	return out, rows.Err()
}
