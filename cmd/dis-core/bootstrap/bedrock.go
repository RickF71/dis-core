package bootstrap

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	identity "dis-core/internal/core/identity"
	discapsule "dis-core/internal/discapsule"
	"dis-core/internal/envx"
)

// NewCapsule creates a Capsule. Tests may override this to inject a mock.
var NewCapsule = func() discapsule.Capsule {
	return discapsule.NewLocalCapsule()
}

// RunBedrockBootstrap performs a one-time, human-approved creation
// of the 1D root domain `null`. If `null` already exists (or an older
// `domain.null` record is present for backwards-compat), this is a no-op.
func RunBedrockBootstrap(ctx context.Context, db *pgxpool.Pool) error {
	log.Println("🪨 BedrockBootstrap: ensuring 1D root domain `null` under human approval")

	// 1) Quick table existence check (same idea as EnsureRootDomain)
	var domainsTableExists bool
	if err := db.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT FROM information_schema.tables
            WHERE table_name = 'domains'
        )
    `).Scan(&domainsTableExists); err != nil {
		return fmt.Errorf("bedrock: domains table existence check failed: %w", err)
	}
	if !domainsTableExists {
		log.Println("   ⚠️  Domains table not found, skipping BedrockBootstrap (DB not provisioned yet)")
		return nil
	}

	// 2) If `null` (canonical) or legacy `domain.null` already exists, bedrock is complete.
	var exists bool
	if err := db.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM domains WHERE name = 'null' OR name = 'domain.null'
        )
    `).Scan(&exists); err != nil {
		return fmt.Errorf("bedrock: root domain existence check failed: %w", err)
	}
	if exists {
		log.Println("   ✅ Root domain already present (`null` or legacy `domain.null`); BedrockBootstrap is a no-op")
		return nil
	}

	// Optional: you could add extra checks here (no seats, no receipts, etc.) if you
	// want to strictly enforce "empty world" before bedrock.

	// 3) Build BedrockChallenge
	ch := discapsule.BedrockChallenge{
		InstanceID: uuid.New().String(),
		Nonce:      uuid.New().String(),
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	// 4) Call capsule for human approval. Use NewCapsule so tests can inject a mock.
	var grant discapsule.BedrockGrant
	if envx.InTestBootstrap() {
		// Non-interactive test bootstrap: auto-approve the grant so tests and
		// CI won't block on stdin prompts.
		log.Println("   ℹ️  DIS_TEST_BOOTSTRAP=1 detected; auto-approving Bedrock grant (non-interactive)")
		grant = discapsule.BedrockGrant{GrantID: uuid.NewString(), Approved: true}
	} else {
		capsule := NewCapsule()
		var err error
		grant, err = capsule.PerformBedrockAuth(ch)
		if err != nil {
			return fmt.Errorf("bedrock: not authorized by capsule/human: %w", err)
		}
	}

	// 5) Run the bedrock creation atomic sequence
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("bedrock: failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := identity.KnowThyselfBedrock(ctx, tx, grant); err != nil {
		return fmt.Errorf("bedrock: KnowThyselfBedrock failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("bedrock: commit failed: %w", err)
	}

	log.Println("   ✅ BedrockBootstrap complete: structural root `null` established")
	return nil
}
