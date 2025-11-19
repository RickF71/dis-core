package bootstrap

import (
	"context"
	"fmt"
	"log"

	"dis-core/internal/identity"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BootstrapIdentityTriads ensures all identities have terra/numen/lima seats (GOV-1)
func BootstrapIdentityTriads(ctx context.Context, db *pgxpool.Pool) error {
	log.Println("🌍 Bootstrapping identity triads (terra/numen/lima) - GOV-1...")

	triadRepo := identity.NewTriadRepository(db)

	// Get all identities
	identityIDs, err := triadRepo.GetAllIdentities(ctx)
	if err != nil {
		return fmt.Errorf("failed to get identities: %w", err)
	}

	if len(identityIDs) == 0 {
		log.Println("   No identities found, skipping triad bootstrap")
		return nil
	}

	// Initialize triads for all identities
	created := 0
	failed := 0

	for _, identityID := range identityIDs {
		triad, err := triadRepo.InitializeTriad(ctx, identityID)
		if err != nil {
			log.Printf("   ⚠️  Failed to initialize triad for %s: %v", identityID, err)
			failed++
			continue
		}

		// Check if seats were newly created (have recent creation time)
		if triad.Terra != nil {
			created++
		}
	}

	log.Printf("✅ Identity triads: %d identities, %d triads created/verified, %d failed",
		len(identityIDs), created, failed)

	// Report on missing triads
	missing, err := triadRepo.GetMissingTriads(ctx)
	if err != nil {
		log.Printf("   ⚠️  Could not check for missing triads: %v", err)
	} else if len(missing) > 0 {
		log.Printf("   ⚠️  %d identities with incomplete triads", len(missing))
	}

	return nil
}
