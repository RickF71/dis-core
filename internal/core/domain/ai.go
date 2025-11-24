package domain

import (
    "context"
    "fmt"
    "strings"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

// ValidateAIDomainPlacementTx validates that an AI domain (child) is placed
// correctly under a user domain which itself is under a corporeal domain.
func ValidateAIDomainPlacementTx(ctx context.Context, tx pgx.Tx, parentID string) error {
    if parentID == "" {
        return fmt.Errorf("ai domain requires a parent user domain")
    }

    var parentType, parentName, grandParentID string
    if err := tx.QueryRow(ctx, `SELECT COALESCE(domain_type,''), COALESCE(name,''), COALESCE(parent_id::text,'') FROM domains WHERE id = $1 LIMIT 1`, parentID).Scan(&parentType, &parentName, &grandParentID); err != nil {
        return fmt.Errorf("validate ai placement: parent lookup: %w", err)
    }

    // If parent is explicitly a corporeal domain, reject immediately.
    if parentType == "corporeal" || strings.Contains(parentName, "corporeal") {
        return fmt.Errorf("ai domains may only be children of user domains")
    }

    // Consider explicit user domain types or name patterns.
    if parentType == "user" || strings.Contains(parentName, "domain.user.") || strings.Contains(parentName, ".user.") {
        // Verify grandparent exists and is corporeal (or named as such).
        if grandParentID == "" {
            return fmt.Errorf("ai domain parent (user) must be under a corporeal domain")
        }
        var gpType, gpName string
        if err := tx.QueryRow(ctx, `SELECT COALESCE(domain_type,''), COALESCE(name,'') FROM domains WHERE id = $1 LIMIT 1`, grandParentID).Scan(&gpType, &gpName); err != nil {
            return fmt.Errorf("validate ai placement: grandparent lookup: %w", err)
        }
        if gpType == "corporeal" || strings.Contains(gpName, "corporeal") {
            return nil
        }
        return fmt.Errorf("ai domain parent (user) must be a child of a corporeal domain")
    }

    return fmt.Errorf("ai domains must be created as children of a user domain")
}

// ValidateAIDomainPlacementPool is a convenience wrapper for callers holding
// a pgxpool.Pool instead of a transactional pgx.Tx.
func ValidateAIDomainPlacementPool(ctx context.Context, pool *pgxpool.Pool, parentID string) error {
    if parentID == "" {
        return fmt.Errorf("ai domain requires a parent user domain")
    }
    tx, err := pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("validate ai placement: begin tx: %w", err)
    }
    defer tx.Rollback(ctx) // safe no-op if committed
    return ValidateAIDomainPlacementTx(ctx, tx, parentID)
}
