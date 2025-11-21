package policy

import (
	"context"

	"dis-core/internal/core/identity"

	"github.com/jackc/pgx/v5"
)

// BuildContextTx loads actor/domain/session context data within the provided
// transaction and returns a map suitable for passing to policy evaluation input.
func BuildContextTx(ctx context.Context, tx pgx.Tx, actorID string, domainID string) (map[string]interface{}, error) {
	actx, err := identity.LoadActorContextTx(ctx, tx, actorID, domainID)
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{
		"actor_id":     actx.ActorID,
		"domain_id":    actx.DomainID,
		"seat_id":      actx.SeatID,
		"presentation": actx.PresentationName,
		"domain_fdn":   actx.DomainFDN,
		"permissions":  actx.Permissions,
		"roles":        actx.Roles,
	}
	return m, nil
}
