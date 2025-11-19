package identity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrNoCorporealDomain = errors.New("identity has no corporeal domain")
)

// IdentityStore abstracts DB access; adapt to your existing store struct.
type IdentityStore struct {
	DB *sql.DB
}

// ResolveCorporealDomain returns the corporeal domain for an identity.
func (s *IdentityStore) ResolveCorporealDomain(ctx context.Context, identityID string) (string, error) {
	const q = `
        SELECT corporeal_domain_id
        FROM identities
        WHERE id = $1
    `
	var domainID sql.NullString
	if err := s.DB.QueryRowContext(ctx, q, identityID).Scan(&domainID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("identity %s not found: %w", identityID, err)
		}
		return "", fmt.Errorf("resolve corporeal domain: %w", err)
	}
	if !domainID.Valid || domainID.String == "" {
		return "", ErrNoCorporealDomain
	}
	return domainID.String, nil
}
