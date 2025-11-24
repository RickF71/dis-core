package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// GetBootstrapActorID returns the configured bootstrap actor_id if present.
func GetBootstrapActorID(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}) (string, error) {
	// First, check safely whether the table exists to avoid raising an error
	// that would abort a surrounding transaction.
	var exists bool
	if err := q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'bootstrap_state')`).Scan(&exists); err != nil {
		return "", fmt.Errorf("get bootstrap actor id (existence check): %w", err)
	}
	if !exists {
		return "", nil
	}

	var actorID *string
	if err := q.QueryRow(ctx, `SELECT actor_id FROM bootstrap_state LIMIT 1`).Scan(&actorID); err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("get bootstrap actor id: %w", err)
	}
	if actorID == nil {
		return "", nil
	}
	return *actorID, nil
}

// SetBootstrapActorIDTx sets the bootstrap actor id only if currently empty.
func SetBootstrapActorIDTx(ctx context.Context, tx pgx.Tx, actorID string) error {
	if tx == nil {
		return fmt.Errorf("SetBootstrapActorIDTx: tx is nil")
	}
	// Ensure the table exists before attempting an UPDATE to avoid a
	// transaction-aborting error when the table is missing.
	if _, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS bootstrap_state (actor_id TEXT)`); err != nil {
		return fmt.Errorf("set bootstrap actor id: create table: %w", err)
	}

	// Seed a single row if none exists so the subsequent UPDATE has a target.
	if _, err := tx.Exec(ctx, `INSERT INTO bootstrap_state (actor_id) SELECT NULL WHERE NOT EXISTS (SELECT 1 FROM bootstrap_state)`); err != nil {
		return fmt.Errorf("set bootstrap actor id: seed row: %w", err)
	}

	// Attempt to set only when current value is NULL or empty
	if _, err := tx.Exec(ctx, `UPDATE bootstrap_state SET actor_id = $1 WHERE actor_id IS NULL OR actor_id = ''`, actorID); err != nil {
		return fmt.Errorf("set bootstrap actor id: %w", err)
	}
	return nil
}
