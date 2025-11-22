package identity

import (
	"context"
	"fmt"
	"time"

	discapsule "dis-core/internal/discapsule"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// KnowThyselfBedrock performs the minimal atomic steps to establish
// the 1D structural root `null`. It is called exactly once, under
// human/capsule approval, from RunBedrockBootstrap.
//
// NOTE: This function intentionally focuses on domain creation. The
// proto-actor, pseat, and receipt logic can be fleshed out once we
// align exactly with the current identity/seat/receipt schemas.
func KnowThyselfBedrock(ctx context.Context, tx pgx.Tx, grant discapsule.BedrockGrant) error {
	if !grant.Approved {
		return fmt.Errorf("bedrock grant not approved")
	}

	// Safety: ensure we don't double-create `null` inside this tx.
	var exists bool
	if err := tx.QueryRow(ctx, `
        SELECT EXISTS (SELECT 1 FROM domains WHERE name = 'null' OR name = 'domain.null')
    `).Scan(&exists); err != nil {
		return fmt.Errorf("bedrock: root domain existence recheck failed: %w", err)
	}
	if exists {
		// Someone else created it in parallel (extremely unlikely, but safe)
		return nil
	}

	// Create domain `null`. We mirror the shape from EnsureRootDomain,
	// but use the canonical name `null` instead of `domain.null`.
	now := time.Now().UTC()
	id := uuid.New()

	if _, err := tx.Exec(ctx, `
        INSERT INTO domains (id, name, parent_id, payload, created_at)
        VALUES ($1, $2, $3, '{}'::jsonb, $4)
    `, id, "null", nil, now); err != nil {
		return fmt.Errorf("bedrock: failed to create root domain `null`: %w", err)
	}

	// TODO (you / later phase):
	//  - create proto-actor (adam) as a structural identity
	//  - create null.pseat bound to that proto-actor
	//  - emit a ci.call.v1 receipt "bedrock.create.v1" linking to grant.GrantID
	//
	// This is where your full-blown Genesis/Bedrock receipts story will plug in.

	return nil
}
