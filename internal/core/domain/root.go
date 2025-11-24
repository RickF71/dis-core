package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// HasRootSovereignTx checks whether the null domain already has a prime/pseat
// assigned that would act as the root sovereign. It expects the caller to
// provide the null domain id and an active transaction so reads see uncommitted
// inserts in the same tx.
func HasRootSovereignTx(ctx context.Context, tx pgx.Tx, nullDomainID uuid.UUID) (bool, error) {
	var sid string
	err := tx.QueryRow(ctx, `SELECT id::text FROM domain_seats WHERE domain_id = $1::uuid AND seat_type = 'prime' LIMIT 1`, nullDomainID.String()).Scan(&sid)
	if err == nil {
		return true, nil
	}
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return false, fmt.Errorf("has root sovereign query: %w", err)
}
