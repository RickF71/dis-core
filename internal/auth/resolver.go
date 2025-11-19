package auth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CorporealDomainBinding represents a resolved corporeal domain from an external UID
type CorporealDomainBinding struct {
	Found              bool
	CorporealDomainID  int64
	CorporealDomainUID string
}

// ResolveCorporealDomainByExternalUID maps an external UID to a corporeal domain
// via the identity_bindings table.
//
// SECURITY: This function ONLY queries - it never stores or logs the external UID.
// The external UID exists only in memory during the request lifecycle.
func ResolveCorporealDomainByExternalUID(ctx context.Context, db *pgxpool.Pool, externalUID string) (CorporealDomainBinding, error) {
	if externalUID == "" {
		return CorporealDomainBinding{Found: false}, nil
	}

	var result CorporealDomainBinding

	// Query identity_bindings to find corporeal domain
	// Note: This assumes identity_bindings.uid contains the external UID
	// and identity_bindings.domain references the corporeal domain ID
	query := `
		SELECT
			ib.domain,
			d.id
		FROM identity_bindings ib
		LEFT JOIN domains d ON d.name = ib.domain OR d.id::text = ib.domain
		WHERE ib.uid = $1
		LIMIT 1
	`

	var domainUID string
	var domainID *string

	err := db.QueryRow(ctx, query, externalUID).Scan(&domainUID, &domainID)

	if err != nil {
		// pgx.ErrNoRows means no binding exists - this is not an error
		if err.Error() == "no rows in result set" {
			return CorporealDomainBinding{Found: false}, nil
		}
		return CorporealDomainBinding{Found: false}, fmt.Errorf("failed to resolve corporeal domain: %w", err)
	}

	// Parse domain ID if available
	var parsedDomainID int64
	if domainID != nil && *domainID != "" {
		// Try to parse as int64 (if domains.id is numeric)
		// If it's UUID-based, we'll use the UID string instead
		fmt.Sscanf(*domainID, "%d", &parsedDomainID)
	}

	result.Found = true
	result.CorporealDomainUID = domainUID
	result.CorporealDomainID = parsedDomainID

	return result, nil
}
