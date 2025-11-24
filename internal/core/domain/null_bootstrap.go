package domain

import (
	"context"
	"fmt"
	"time"

	"dis-core/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// EnsureNullDomainExistsTx ensures a domain row named 'domain.null' exists
// and returns its UUID. Caller provides an active pgx.Tx.
func EnsureNullDomainExistsTx(ctx context.Context, tx pgx.Tx) (uuid.UUID, error) {
	const name = "domain.null"
	var existing string
	if err := tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = $1 LIMIT 1`, name).Scan(&existing); err == nil {
		id, _ := uuid.Parse(existing)
		return id, nil
	}

	id := uuid.New()
	now := time.Now().UTC()
	if _, err := tx.Exec(ctx, `INSERT INTO domains (id, name, parent_id, payload, created_at) VALUES ($1,$2,NULL,'{}'::jsonb,$3)`, id.String(), name, now); err != nil {
		return uuid.Nil, fmt.Errorf("ensure null domain insert: %w", err)
	}
	return id, nil
}

// EnsureNullPrimeSeatExistsTx ensures a prime seat for the null domain exists
// and is assigned to the provided member (identity). If a matching prime
// seat already exists and is assigned to this member, it's a no-op. Returns
// the seat id.
func EnsureNullPrimeSeatExistsTx(ctx context.Context, tx pgx.Tx, domainID uuid.UUID, memberID string) (uuid.UUID, error) {
	// If any prime seat already exists for this domain, refuse to silently
	// create another: return the existing seat if it's assigned to the
	// requested member, otherwise return an error to block recreation.
	var sid string
	var existingMember string
	if err := tx.QueryRow(ctx, `SELECT id::text, COALESCE(member_id,'') FROM domain_seats WHERE domain_id = $1::uuid AND seat_type = 'prime' LIMIT 1`, domainID.String()).Scan(&sid, &existingMember); err == nil {
		if existingMember == memberID {
			id, _ := uuid.Parse(sid)
			return id, nil
		}
		return uuid.Nil, fmt.Errorf("prime seat already exists for domain %s assigned to %s", domainID.String(), existingMember)
	}

	seatID := uuid.New()
	if err := db.InsertSeatTx(ctx, tx, seatID.String(), domainID.String(), memberID, "prime"); err != nil {
		return uuid.Nil, fmt.Errorf("ensure null prime seat insert: %w", err)
	}
	return seatID, nil
}
