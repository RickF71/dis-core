# /task Add BedrockBootstrap (structural genesis) to DIS-Core
# Objective: Replace EnsureRootDomain with a capsule-authorized BedrockBootstrap step.
# This will create the 1D root domain `null` only once under human approval.

# Changes:
# 1. Modify cmd/dis-core/main.go to call RunBedrockBootstrap instead of EnsureRootDomain.
# 2. Add new file cmd/dis-core/bootstrap/bedrock.go containing the BedrockBootstrap orchestrator.
# 3. Add new folder internal/discapsule with:
#       - capsule.go (interface + BedrockChallenge and BedrockGrant)
#       - local_capsule.go (interactive capsule implementation)
# 4. Add new file internal/core/identity/know_thyself_bedrock.go for minimal bedrock creation.
# 5. Do not delete EnsureRootDomain yet; just stop calling it.

=== modify: cmd/dis-core/main.go ===
Replace the EnsureRootDomain block with:
    if err := bootstrap.RunBedrockBootstrap(ctx, dbComponents.Database); err != nil {
        log.Fatalf("bedrock bootstrap failed: %v", err)
    }

=== create: cmd/dis-core/bootstrap/bedrock.go ===
package bootstrap

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"

    discapsule "dis-core/internal/discapsule"
    identity "dis-core/internal/core/identity"
)

func RunBedrockBootstrap(ctx context.Context, db *pgxpool.Pool) error {
    log.Println("🪨 BedrockBootstrap: ensuring root domain `null` under human approval")

    var tableExists bool
    if err := db.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT FROM information_schema.tables
            WHERE table_name = 'domains'
        )
    `).Scan(&tableExists); err != nil {
        return fmt.Errorf("bedrock: domains table existence check failed: %w", err)
    }
    if !tableExists {
        log.Println("   ⚠️ domains table missing — skipping bedrock")
        return nil
    }

    var exists bool
    if err := db.QueryRow(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM domains WHERE name = 'null' OR name = 'domain.null'
        )
    `).Scan(&exists); err != nil {
        return fmt.Errorf("bedrock: root domain existence check failed: %w", err)
    }
    if exists {
        log.Println("   ✅ root domain exists; bedrock is complete")
        return nil
    }

    ch := discapsule.BedrockChallenge{
        InstanceID: uuid.New().String(),
        Nonce:      uuid.New().String(),
        Timestamp:  time.Now().UTC().Format(time.RFC3339),
    }

    capsule := discapsule.NewLocalCapsule()
    grant, err := capsule.PerformBedrockAuth(ch)
    if err != nil {
        return fmt.Errorf("bedrock: not authorized: %w", err)
    }

    tx, err := db.Begin(ctx)
    if err != nil {
        return fmt.Errorf("bedrock: begin tx failed: %w", err)
    }
    defer tx.Rollback(ctx)

    if err := identity.KnowThyselfBedrock(ctx, tx, grant); err != nil {
        return fmt.Errorf("bedrock: KnowThyselfBedrock failed: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("bedrock: commit failed: %w", err)
    }

    log.Println("   ✅ BedrockBootstrap complete")
    return nil
}

=== create: internal/discapsule/capsule.go ===
package discapsule

type BedrockChallenge struct {
    InstanceID string `json:"instance_id"`
    Nonce      string `json:"nonce"`
    Timestamp  string `json:"timestamp"`
}

type BedrockGrant struct {
    GrantID    string `json:"grant_id"`
    InstanceID string `json:"instance_id"`
    Nonce      string `json:"nonce"`
    Approved   bool   `json:"approved"`
}

type Capsule interface {
    PerformBedrockAuth(ch BedrockChallenge) (BedrockGrant, error)
}

=== create: internal/discapsule/local_capsule.go ===
package discapsule

import (
    "bufio"
    "fmt"
    "os"
    "strings"

    "github.com/google/uuid"
)

type LocalCapsule struct{}

func NewLocalCapsule() *LocalCapsule {
    return &LocalCapsule{}
}

func (c *LocalCapsule) PerformBedrockAuth(ch BedrockChallenge) (BedrockGrant, error) {
    fmt.Println("🪨 DIS-Core BedrockBootstrap")
    fmt.Println("  DIS-Core requests approval to create the 1D root domain `null`.")
    fmt.Printf("  Instance: %s\n  Nonce:    %s\n  Time:     %s\n", ch.InstanceID, ch.Nonce, ch.Timestamp)
    fmt.Print("  Approve bedrock creation? (y/n): ")

    reader := bufio.NewReader(os.Stdin)
    line, _ := reader.ReadString('\n')
    line = strings.TrimSpace(line)

    if line != "y" && line != "Y" {
        return BedrockGrant{}, fmt.Errorf("bedrock declined by human")
    }

    return BedrockGrant{
        GrantID:    uuid.New().String(),
        InstanceID: ch.InstanceID,
        Nonce:      ch.Nonce,
        Approved:   true,
    }, nil
}

=== create: internal/core/identity/know_thyself_bedrock.go ===
package identity

import (
    "context"
    "fmt"
    "time"

    discapsule "dis-core/internal/discapsule"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
)

func KnowThyselfBedrock(ctx context.Context, tx pgx.Tx, grant discapsule.BedrockGrant) error {
    if !grant.Approved {
        return fmt.Errorf("bedrock: grant not approved")
    }

    var exists bool
    if err := tx.QueryRow(ctx, `
        SELECT EXISTS (SELECT 1 FROM domains WHERE name = 'null' OR name = 'domain.null')
    `).Scan(&exists); err != nil {
        return fmt.Errorf("bedrock: existence recheck failed: %w", err)
    }
    if exists {
        return nil
    }

    id := uuid.New()
    now := time.Now().UTC()

    _, err := tx.Exec(ctx, `
        INSERT INTO domains (id, name, parent_id, payload, created_at)
        VALUES ($1, 'null', NULL, '{}'::jsonb, $2)
    `, id, now)
    if err != nil {
        return fmt.Errorf("bedrock: failed to create `null`: %w", err)
    }

    // TODO:
    //  - create proto-actor (adam)
    //  - create null.pseat
    //  - write bedrock receipt (ci.call.v1)
    return nil
}

