package bootstrap

import (
	"context"
	"fmt"
	"log"

	"dis-core/internal/seats"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PhaseS0PrimeSeatSetup ensures every domain has a Prime Seat
func PhaseS0PrimeSeatSetup(ctx context.Context, db *pgxpool.Pool) error {
	log.Println("🧱 Phase S0 — Ensuring Prime Seats for all domains...")

	// Get all domain IDs
	rows, err := db.Query(ctx, "SELECT id FROM domains")
	if err != nil {
		return fmt.Errorf("query domains: %w", err)
	}
	defer rows.Close()

	var domainIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("scan domain ID: %w", err)
		}
		domainIDs = append(domainIDs, id)
	}

	// Initialize seats repository
	seatsRepo := seats.NewRepository(db)

	// Ensure Prime Seat for each domain
	created := 0
	for _, domainID := range domainIDs {
		_, err := seatsRepo.EnsurePrimeSeat(ctx, domainID)
		if err != nil {
			log.Printf("  Warning: Could not ensure Prime Seat for domain %s: %v", domainID, err)
			continue
		}
		created++
	}

	log.Printf("✅ Phase S0 — Prime Seats ensured (%d domains processed, %d seats created/verified)", len(domainIDs), created)
	return nil
}
