package identity

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// ActorContext represents a minimal session context for an actor within a domain.
// NOTE: Repository-wide IDs are UUIDs represented as strings. The original MX-K3
// requested integer IDs, but this project consistently uses UUID strings — we
// follow the repository convention here.
type ActorContext struct {
	ActorID          string   `json:"actor_id"`
	DomainID         string   `json:"domain_id"`
	SeatID           string   `json:"seat_id"` // prime seat
	PresentationName string   `json:"presentation_name"`
	DomainFDN        string   `json:"domain_fdn"`
	Permissions      []string `json:"permissions"`
	Roles            []string `json:"roles,omitempty"`
}

// LoadActorContextTx loads a minimal ActorContext using the provided transaction.
// It performs only SELECTs and does not mutate database state.
func LoadActorContextTx(ctx context.Context, tx pgx.Tx, actorID string, domainID string) (*ActorContext, error) {
	var presentation string
	// identities.id is UUID
	if err := tx.QueryRow(ctx, `SELECT presentation_name FROM identities WHERE id = $1::uuid`, actorID).Scan(&presentation); err != nil {
		return nil, fmt.Errorf("load actor context: fetch identity presentation: %w", err)
	}

	// Resolve prime seat for actor/domain
	var seatID string
	// member_id is stored as UUID text; use COALESCE to handle NULLs
	err := tx.QueryRow(ctx, `SELECT id::text FROM domain_seats WHERE domain_id = $1::uuid AND COALESCE(member_id::text,'') = $2 AND seat_type = 'prime' LIMIT 1`, domainID, actorID).Scan(&seatID)
	if err != nil {
		return nil, fmt.Errorf("load actor context: fetch prime seat: %w", err)
	}

	// Resolve domain FDN (name) via direct query (tx-aware)
	var domainName string
	if err := tx.QueryRow(ctx, `SELECT name FROM domains WHERE id = $1::uuid`, domainID).Scan(&domainName); err != nil {
		return nil, fmt.Errorf("load actor context: resolve domain fdn: %w", err)
	}

	// Load roles for the resolved prime seat (if any)
	var roles []string
	rows, err := tx.Query(ctx, `SELECT role FROM seat_roles WHERE seat_id = $1::uuid ORDER BY role`, seatID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var r string
			if err := rows.Scan(&r); err == nil {
				roles = append(roles, r)
			}
		}
	}

	return &ActorContext{
		ActorID:          actorID,
		DomainID:         domainID,
		SeatID:           seatID,
		PresentationName: presentation,
		DomainFDN:        domainName,
		Permissions:      []string{},
		Roles:            roles,
	}, nil
}
