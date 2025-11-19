package representation

import (
	"context"
	"fmt"

	"dis-core/internal/identity"
	"dis-core/internal/seats"
)

// Resolver ties together identity + seats + (later) actors.
type Resolver struct {
	Identities *identity.IdentityStore
	Seats      *seats.SeatStore
	// TODO: add ActorStore when foreign domains are active.
}

func NewResolver(idStore *identity.IdentityStore, seatStore *seats.SeatStore) *Resolver {
	return &Resolver{
		Identities: idStore,
		Seats:      seatStore,
	}
}

// ResolveForDomain determines how an identity appears in targetDomainID.
func (r *Resolver) ResolveForDomain(ctx context.Context, identityID, targetDomainID string) (*Representation, error) {
	corpDomainID, err := r.Identities.ResolveCorporealDomain(ctx, identityID)
	if err != nil {
		return nil, fmt.Errorf("resolve corporeal domain for identity %s: %w", identityID, err)
	}

	// Default target domain to corporeal domain if none is specified.
	if targetDomainID == "" {
		targetDomainID = corpDomainID
	}

	if corpDomainID == targetDomainID {
		seatID, err := r.Seats.GetPrimeSeatForCorporealDomain(ctx, corpDomainID)
		if err != nil {
			return nil, fmt.Errorf("resolve prime seat for domain %s: %w", corpDomainID, err)
		}

		return &Representation{
			Kind:              KindSovereign,
			IdentityID:        identityID,
			CorporealDomainID: corpDomainID,
			DomainID:          targetDomainID,
			SeatID:            seatID,
			ActorID:           "",
		}, nil
	}

	return nil, fmt.Errorf("foreign domain representation not yet implemented (identity=%s targetDomain=%s)", identityID, targetDomainID)
}
