package seats

import (
	"context"
	"database/sql"
	"fmt"
)

type SeatStore struct {
	DB *sql.DB
}

// GetPrimeSeatForCorporealDomain finds the pseat for the given domain.
func (s *SeatStore) GetPrimeSeatForCorporealDomain(ctx context.Context, domainID string) (string, error) {
	const q = `
        SELECT id
        FROM seats
        WHERE domain_id = $1
          AND kind = 'pseat'
          AND active = TRUE
        LIMIT 1
    `
	var seatID string
	if err := s.DB.QueryRowContext(ctx, q, domainID).Scan(&seatID); err != nil {
		return "", fmt.Errorf("get prime seat for domain %s: %w", domainID, err)
	}
	return seatID, nil
}
