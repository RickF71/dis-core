package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GOV-8: DIS-Invariant-001 — No authority, lookup, or receipt may rely on domain names.
// All internal operations must reference domain.id (UUID) as the single canonical identity.
//
// ResolveDomainFDN retrieves the Fully Distinguished Name (FDN) for a domain UUID.
// This is ONLY for human-readable logs, debugging, and display purposes.
// NEVER use this for authorization, lookups, or business logic.
func ResolveDomainFDN(ctx context.Context, pool *pgxpool.Pool, domainID uuid.UUID) (string, error) {
	var name string
	query := `SELECT name FROM domains WHERE id = $1`
	err := pool.QueryRow(ctx, query, domainID).Scan(&name)
	if err != nil {
		return "", fmt.Errorf("GOV-8: failed to resolve FDN for domain %s: %w", domainID, err)
	}
	return name, nil
}

// GOV-8: ResolveDomainLineage builds a UUID-based lineage path for audit trails
// Returns array of UUIDs from root to leaf: [terra, numen, lima, corporeal, user.rick]
func ResolveDomainLineage(ctx context.Context, pool *pgxpool.Pool, domainID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		WITH RECURSIVE lineage AS (
			-- Base case: start with the given domain
			SELECT id, parent_id, 0 as depth
			FROM domains
			WHERE id = $1

			UNION ALL

			-- Recursive case: walk up to parents
			SELECT d.id, d.parent_id, l.depth + 1
			FROM domains d
			INNER JOIN lineage l ON d.id = l.parent_id
			WHERE d.parent_id IS NOT NULL AND l.depth < 20
		)
		SELECT id FROM lineage ORDER BY depth DESC
	`

	rows, err := pool.Query(ctx, query, domainID)
	if err != nil {
		return nil, fmt.Errorf("GOV-8: failed to resolve lineage for domain %s: %w", domainID, err)
	}
	defer rows.Close()

	var lineage []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("GOV-8: failed to scan lineage UUID: %w", err)
		}
		lineage = append(lineage, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GOV-8: lineage iteration error: %w", err)
	}

	return lineage, nil
}

// GOV-8: ResolveDomainLineageFDN builds human-readable lineage path for logs
// Example: ["terra", "numen", "lima", "corporeal", "user.rick"]
// WARNING: This is for display ONLY. Never use for authorization.
func ResolveDomainLineageFDN(ctx context.Context, pool *pgxpool.Pool, domainID uuid.UUID) ([]string, error) {
	uuidLineage, err := ResolveDomainLineage(ctx, pool, domainID)
	if err != nil {
		return nil, err
	}

	var fdnLineage []string
	for _, id := range uuidLineage {
		fdn, err := ResolveDomainFDN(ctx, pool, id)
		if err != nil {
			// If we can't resolve a name, use UUID as fallback for display
			fdnLineage = append(fdnLineage, id.String())
			continue
		}
		fdnLineage = append(fdnLineage, fdn)
	}

	return fdnLineage, nil
}

// GOV-8: ValidateDomainID ensures a UUID exists in the domains table
// This is the ONLY acceptable way to verify domain existence
func ValidateDomainID(ctx context.Context, pool *pgxpool.Pool, domainID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM domains WHERE id = $1)`
	err := pool.QueryRow(ctx, query, domainID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("GOV-8: failed to validate domain ID %s: %w", domainID, err)
	}
	return exists, nil
}

// GOV-8: GetDomainMetadata retrieves canonical UUID-based domain information
// Names are included for display only and MUST NOT be used for lookups
type DomainMetadata struct {
	ID          uuid.UUID              `json:"id"`
	ParentID    *uuid.UUID             `json:"parent_id,omitempty"`
	Name        string                 `json:"name"`         // Display only - NOT canonical
	NameWarning string                 `json:"name_warning"` // GOV-8 reminder
	Governance  string                 `json:"governance"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
}

func GetDomainMetadata(ctx context.Context, pool *pgxpool.Pool, domainID uuid.UUID) (*DomainMetadata, error) {
	var meta DomainMetadata
	var payload interface{}

	query := `
		SELECT id, parent_id, name, payload->'meta'->>'governance', payload
		FROM domains
		WHERE id = $1
	`

	err := pool.QueryRow(ctx, query, domainID).Scan(
		&meta.ID,
		&meta.ParentID,
		&meta.Name,
		&meta.Governance,
		&payload,
	)

	if err != nil {
		return nil, fmt.Errorf("GOV-8: failed to get metadata for domain %s: %w", domainID, err)
	}

	meta.NameWarning = "GOV-8: Name is display metadata only. Use ID for all operations."

	if payload != nil {
		if payloadMap, ok := payload.(map[string]interface{}); ok {
			meta.Payload = payloadMap
		}
	}

	return &meta, nil
}
