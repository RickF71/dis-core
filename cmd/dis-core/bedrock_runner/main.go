package main

import (
	"context"
	"log"

	"dis-core/cmd/dis-core/bootstrap"
	identity "dis-core/internal/core/identity"
	discapsule "dis-core/internal/discapsule"
)

func main() {
	ctx := context.Background()

	cfg := bootstrap.LoadConfig()

	dbComponents, err := bootstrap.InitializeDatabase(ctx, cfg.DSN)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer dbComponents.Close()

	if err := bootstrap.BootstrapTables(dbComponents.Database); err != nil {
		log.Fatalf("failed to bootstrap tables: %v", err)
	}

	// Use an auto-approved grant to run the bedrock creation non-interactively
	grant := discapsule.BedrockGrant{
		GrantID:    "auto-run",
		InstanceID: "auto-run-instance",
		Nonce:      "auto-run-nonce",
		Approved:   true,
	}

	tx, err := dbComponents.Database.Begin(ctx)
	if err != nil {
		log.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback(ctx)

	if err := identity.KnowThyselfBedrock(ctx, tx, grant); err != nil {
		log.Fatalf("KnowThyselfBedrock failed: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("commit failed: %v", err)
	}

	log.Println("Bedrock runner: success — root domain ensured")
}
