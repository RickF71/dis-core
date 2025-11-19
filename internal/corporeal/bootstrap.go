package corporeal

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Bootstrapper handles atomic corporeal + actor-domain creation
type Bootstrapper struct {
	DB *pgxpool.Pool
}

// NewBootstrapper creates a new corporeal bootstrapper
func NewBootstrapper(db *pgxpool.Pool) *Bootstrapper {
	return &Bootstrapper{DB: db}
}

// EnsureCorporealBootstrap atomically creates:
// 1. Corporeal domain (parent: terra domain)
// 2. Corporeal Prime Seat (held by HUMAN, not an actor)
// 3. Default actor-domain (child of corporeal domain)
// 4. Default actor identity
// 5. Actor Prime Seat inside actor-domain
// 6. Emit receipts and commit atomically
func (b *Bootstrapper) EnsureCorporealBootstrap(ctx context.Context, userUID string) (string, string, error) {
	// Start transaction for atomic operations
	tx, err := b.DB.Begin(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now().UTC()

	// ---------------------------
	// 1. Create corporeal domain
	// ---------------------------
	corporealDomainID := uuid.New()
	corporealDomainName := fmt.Sprintf("domain.%s", userUID)

	// Check if a domain with this name already exists
	var existingID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = $1`, corporealDomainName).Scan(&existingID)
	if err == nil {
		// Domain already exists, return existing IDs
		var existingActorDomainID uuid.UUID
		actorDomainName := fmt.Sprintf("domain.actor.%s_corp", userUID)
		err = tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = $1`, actorDomainName).Scan(&existingActorDomainID)
		tx.Rollback(ctx)
		if err == nil {
			return existingID.String(), existingActorDomainID.String(), nil
		}
	}

	// Get parent domain UUID (terra)
	var parentID uuid.UUID
	err = tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = 'terra' LIMIT 1`).Scan(&parentID)
	if err != nil {
		return "", "", fmt.Errorf("parent domain 'terra' not found: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO domains (id, name, parent_id, payload, created_at)
		VALUES ($1, $2, $3, '{}'::jsonb, $4)
	`, corporealDomainID, corporealDomainName, parentID, now)
	if err != nil {
		return "", "", fmt.Errorf("corporeal domain insert failed: %w", err)
	}

	// ---------------------------
	// 2. Create corporeal Prime Seat (held by HUMAN)
	// ---------------------------
	corpPSeatID := uuid.New()
	holderID := fmt.Sprintf("human.%s", userUID)

	// Note: domain_seats table uses UUID for id and domain_id, but TEXT for member_id
	_, err = tx.Exec(ctx, `
		INSERT INTO domain_seats (id, domain_id, seat_type, member_id, status, created_at)
		VALUES ($1, $2, 'prime', $3, 'active', $4)
	`, corpPSeatID, corporealDomainID, holderID, now)
	if err != nil {
		return "", "", fmt.Errorf("corporeal Prime Seat failed: %w", err)
	}

	// ---------------------------
	// 3. Create default actor-domain
	// ---------------------------
	actorDomainID := uuid.New()
	actorDomainName := fmt.Sprintf("domain.actor.%s_corp", userUID)

	_, err = tx.Exec(ctx, `
		INSERT INTO domains (id, name, parent_id, payload, created_at)
		VALUES ($1, $2, $3, '{}'::jsonb, $4)
	`, actorDomainID, actorDomainName, corporealDomainID, now)
	if err != nil {
		return "", "", fmt.Errorf("actor-domain insert failed: %w", err)
	}

	// ---------------------------
	// 4. Create default actor (identity)
	// ---------------------------
	actorDisUID := fmt.Sprintf("actor.%s.corp", userUID)
	actorNamespace := fmt.Sprintf("corporeal.%s", userUID)

	var actorID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO identities (dis_uid, namespace, created_at, active)
		VALUES ($1, $2, $3, TRUE)
		RETURNING id
	`, actorDisUID, actorNamespace, now).Scan(&actorID)
	if err != nil {
		return "", "", fmt.Errorf("actor creation failed: %w", err)
	}

	// ---------------------------
	// 5. Assign actor-domain Prime Seat (operational sovereignty)
	// ---------------------------
	actorPSeatID := uuid.New()

	_, err = tx.Exec(ctx, `
		INSERT INTO domain_seats (id, domain_id, seat_type, member_id, status, created_at)
		VALUES ($1, $2, 'prime', $3, 'active', $4)
	`, actorPSeatID, actorDomainID, actorDisUID, now)
	if err != nil {
		return "", "", fmt.Errorf("actor-domain Prime Seat failed: %w", err)
	}

	// ---------------------------
	// 6. Emit receipts
	// ---------------------------
	receipts := []struct {
		receiptType string
		targetID    string
	}{
		{"domain.create.v1", corporealDomainID.String()},
		{"seat.instantiate.v1", corpPSeatID.String()},
		{"domain.create.actor-domain.v1", actorDomainID.String()},
		{"identity.actor.create.v1", actorDisUID},
		{"seat.instantiate.v1", actorPSeatID.String()},
	}

	for _, r := range receipts {
		_, err = tx.Exec(ctx, `
			INSERT INTO receipts (receipt_type, event_id, issued_by, issued_at, verified, metadata)
			VALUES (
				$1,
				$2,
				'system.corporeal-bootstrap',
				$3,
				TRUE,
				jsonb_build_object(
					'user_uid', $4::text,
					'corporeal_domain', $5::text,
					'actor_domain', $6::text,
					'bootstrap_timestamp', $7::text
				)
			)
		`, r.receiptType, r.targetID, now, userUID, corporealDomainID.String(), actorDomainID.String(), now.Format(time.RFC3339))
		if err != nil {
			return "", "", fmt.Errorf("failed to emit receipt %s: %w", r.receiptType, err)
		}
	}

	// ---------------------------
	// 7. Commit transaction
	// ---------------------------
	if err := tx.Commit(ctx); err != nil {
		return "", "", fmt.Errorf("commit failed: %w", err)
	}

	return corporealDomainID.String(), actorDomainID.String(), nil
}
