package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// GetDomainTx fetches minimal domain metadata inside a transaction.
func GetDomainTx(ctx context.Context, tx pgx.Tx, domainID string) (map[string]interface{}, error) {
	row := tx.QueryRow(ctx, `SELECT id::text, name FROM domains WHERE id = $1::uuid`, domainID)
	var id, name string
	if err := row.Scan(&id, &name); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get domain tx: %w", err)
	}
	return map[string]interface{}{"id": id, "name": name}, nil
}
