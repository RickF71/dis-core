package bedrock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dis-core/internal/core/authority"
	"dis-core/internal/core/domain/spineconfig"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// BedrockBootstrapAtomic runs inside an existing tx and ensures a minimal
// canonical spine exists. This implementation is intentionally small and
// idempotent so tests can run while we keep the contract stable.
func BedrockBootstrapAtomic(ctx context.Context, tx pgx.Tx) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM domains WHERE name = 'null' OR name = 'domain.null')`).Scan(&exists); err != nil {
		return fmt.Errorf("bedrock: existence check: %w", err)
	}
	if exists {
		return nil
	}

	now := time.Now().UTC()

	// insert null root
	if _, err := tx.Exec(ctx, `INSERT INTO domains (id, name, parent_id, payload, created_at) VALUES ($1,$2,NULL,'{}'::jsonb,$3) ON CONFLICT (name) DO NOTHING`, uuid.New().String(), "null", now); err != nil {
		return fmt.Errorf("bedrock: insert null: %w", err)
	}

	// ensure aether exists as a direct child of null (idempotent)
	if _, err := tx.Exec(ctx, `INSERT INTO domains (id, name, parent_id, payload, created_at) VALUES (gen_random_uuid(), 'aether', (SELECT id FROM domains WHERE name = 'null' OR name = 'domain.null' LIMIT 1), '{}', $1) ON CONFLICT (name) DO NOTHING`, now); err != nil {
		return fmt.Errorf("bedrock: insert aether: %w", err)
	}

	// resolve current parent id (starts with null)
	var parentID string
	if err := tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = 'null' OR name = 'domain.null' LIMIT 1`).Scan(&parentID); err != nil {
		return fmt.Errorf("bedrock: resolve null id: %w", err)
	}

	// helper to insert a named child domain with the given parent; idempotent
	insertChild := func(name string, p string) error {
		id := uuid.New().String()
		if _, err := tx.Exec(ctx, `INSERT INTO domains (id, name, parent_id, payload, created_at) VALUES ($1,$2,$3,'{}'::jsonb,$4) ON CONFLICT (name) DO NOTHING`, id, name, p, now); err != nil {
			return fmt.Errorf("bedrock: insert %s: %w", name, err)
		}
		return nil
	}

	// create void -> null
	if err := insertChild("void", parentID); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = 'void' OR name = 'domain.void' LIMIT 1`).Scan(&parentID); err != nil {
		return fmt.Errorf("bedrock: resolve void id: %w", err)
	}

	// terra -> void
	if err := insertChild("terra", parentID); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = 'terra' OR name = 'domain.terra' LIMIT 1`).Scan(&parentID); err != nil {
		return fmt.Errorf("bedrock: resolve terra id: %w", err)
	}

	// numen -> terra
	if err := insertChild("numen", parentID); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = 'numen' OR name = 'domain.numen' LIMIT 1`).Scan(&parentID); err != nil {
		return fmt.Errorf("bedrock: resolve numen id: %w", err)
	}

	// lima -> numen
	if err := insertChild("lima", parentID); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = 'lima' OR name = 'domain.lima' LIMIT 1`).Scan(&parentID); err != nil {
		// domain.lima fallback not strictly necessary; just ensure exists
	}

	// Emit a bedrock init receipt attached to the null domain
	var nullDomainID string
	if err := tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = 'null' OR name = 'domain.null' LIMIT 1`).Scan(&nullDomainID); err != nil {
		return fmt.Errorf("bedrock: resolve null for receipt: %w", err)
	}
	// Build spine array from canonical spineconfig
	spine := make([]string, 0, len(spineconfig.CanonicalSpine()))
	for _, e := range spineconfig.CanonicalSpine() {
		spine = append(spine, e.Name)
	}
	payload := map[string]any{"spine": spine}
	if _, err := authority.RecordAuthorityReceiptTx(ctx, tx, nullDomainID, "ci.call.bedrock.init.v1", payload); err != nil {
		return fmt.Errorf("bedrock: record receipt: %w", err)
	}

	if err := SetSystemBootstrapped(ctx, tx); err != nil {
		return fmt.Errorf("bedrock: set bootstrapped: %w", err)
	}

	return nil
}

// IsSystemBootstrapped reads the bootstrapped flag from system_state.
func IsSystemBootstrapped(ctx context.Context, tx pgx.Tx) (bool, error) {
	// First check whether the table exists using pg_catalog so we don't
	// attempt to SELECT from a missing table (that would error and put the
	// surrounding transaction into an aborted state). Querying pg_catalog
	// returns a harmless false if the table is absent.
	var tableExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_class WHERE relname = 'system_state' AND relkind = 'r')`).Scan(&tableExists); err != nil {
		return false, fmt.Errorf("is_system_bootstrapped: table existence check: %w", err)
	}
	if !tableExists {
		return false, nil
	}

	var val string
	if err := tx.QueryRow(ctx, `SELECT value FROM system_state WHERE key = 'bootstrapped' LIMIT 1`).Scan(&val); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("is_system_bootstrapped: %w", err)
	}
	return val == "1", nil
}

// SetSystemBootstrapped upserts the bootstrapped flag.
func SetSystemBootstrapped(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS system_state (key TEXT PRIMARY KEY, value TEXT, updated_at TIMESTAMPTZ)`); err != nil {
		return fmt.Errorf("set bootstrapped: create table: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO system_state (key, value, updated_at) VALUES ('bootstrapped','1', now()) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`); err != nil {
		return fmt.Errorf("set bootstrapped: upsert: %w", err)
	}
	return nil
}
