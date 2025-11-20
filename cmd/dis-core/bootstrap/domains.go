package bootstrap

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GOV-6: Bootstrap auto.corporeal synthetic domain
// RegisterSyntheticDomains ensures the auto.corporeal test domain exists
func RegisterSyntheticDomains(ctx context.Context, db *pgxpool.Pool) error {
	log.Println("🤖 GOV-6: Registering synthetic domains...")

	// Check if domains table exists
	var tableExists bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'domains'
		)
	`).Scan(&tableExists)
	if err != nil {
		return err
	}

	if !tableExists {
		log.Println("   ⚠️  Domains table not found, skipping synthetic domain registration")
		return nil
	}

	// Check if auto.corporeal already exists
	var exists bool
	err = db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM domains WHERE name = 'auto.corporeal'
		)
	`).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		log.Println("   ✅ auto.corporeal domain already registered")
		return nil
	}

	// Get simula parent domain ID (or use NULL if not found)
	var parentID *string
	err = db.QueryRow(ctx, `
		SELECT id FROM domains WHERE name = 'simula' LIMIT 1
	`).Scan(&parentID)
	if err != nil {
		log.Println("   ⚠️  Parent domain 'simula' not found, using NULL parent")
		parentID = nil
	}

	// Prepare metadata
	metadata := map[string]interface{}{
		"synthetic":          true,
		"corporeal_warning":  "This domain is not a true corporeal entity.",
		"deterministic_test": true,
		"sandbox":            true,
		"capabilities": map[string]bool{
			"delegation":      false,
			"freeze_override": false,
			"break_glass":     false,
			"authorize":       false,
		},
	}
	metadataJSON, _ := json.Marshal(metadata)

	// Insert auto.corporeal domain
	_, err = db.Exec(ctx, `
		INSERT INTO domains (name, type, parent_id, state, description, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, "auto.corporeal", "corporeal", parentID, "active",
		"Synthetic corporeal domain for deterministic testing.",
		metadataJSON)

	if err != nil {
		return err
	}

	log.Println("   ✅ Created domain: auto.corporeal (synthetic test domain)")
	log.Println("      - Type: corporeal (synthetic)")
	log.Println("      - State: active")
	log.Println("      - Capabilities: sandbox-restricted (no delegation/freeze override)")
	log.Println("      - Warning: Not a true corporeal entity")

	return nil
}

// EnsureRootDomain ensures the special root domain `domain.null` exists.
// This domain is used as the canonical root for deterministic operations and
// must be present at process start. The function is idempotent.
func EnsureRootDomain(ctx context.Context, db *pgxpool.Pool) error {
	log.Println("🔧 Ensuring root domain: domain.null")

	// Quick table existence check to avoid noisy errors when DB isn't provisioned.
	var tableExists bool
	if err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'domains'
		)
	`).Scan(&tableExists); err != nil {
		return err
	}
	if !tableExists {
		log.Println("   ⚠️  Domains table not found, skipping root domain ensure")
		return nil
	}

	// Check if domain.null already exists
	var exists bool
	if err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM domains WHERE name = 'domain.null')`).Scan(&exists); err != nil {
		return err
	}
	if exists {
		log.Println("   ✅ domain.null already present")
		return nil
	}

	// Insert domain.null
	id := uuid.New()
	now := time.Now().UTC()
	_, err := db.Exec(ctx, `
		INSERT INTO domains (id, name, parent_id, payload, created_at)
		VALUES ($1, $2, $3, '{}'::jsonb, $4)
	`, id, "domain.null", nil, now)
	if err != nil {
		return err
	}
	log.Println("   ✅ Created root domain: domain.null")
	return nil
}
