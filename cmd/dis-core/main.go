package main

import (
	"context"
	"log"

	"dis-core/cmd/dis-core/bootstrap"
	"dis-core/cmd/dis-core/service"
	logutil "dis-core/internal/log"
	"dis-core/internal/policy"
)

func main() {
	// Setup logging for daemon/systemd
	logutil.SetupLogging()

	ctx := context.Background()

	// Load configuration
	config := bootstrap.LoadConfig()

	// Initialize database components (connections, registry, ledger, etc.)
	dbComponents, err := bootstrap.InitializeDatabase(ctx, config.DSN)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer dbComponents.Close()

	// Bootstrap database tables (schemas for domains, seats, receipts, etc.)
	if err := bootstrap.BootstrapTables(dbComponents.Database); err != nil {
		log.Printf("Warning: failed to bootstrap tables: %v", err)
	}

	// ✅ Minimal structural bootstrap: ensure domain.null exists.
	// This is the ONLY domain that must exist at process start.
	if err := bootstrap.EnsureRootDomain(ctx, dbComponents.Database); err != nil {
		log.Fatalf("failed to ensure root domain.null: %v", err)
	}

	// ❌ Removed:
	// - PhaseS0PrimeSeatSetup (pseats should be ensured by the domain loader per-branch)
	// - BootstrapIdentityTriads (terra/numen/lima should be adoptable, not globally pre-created)
	// - RegisterSyntheticDomains (synthetic/system domains should be opt-in, not auto-injected)

	// Initialize policy engine (OPA/Rego) – domain-aware policies will be loaded here.
	// You’ll implement this to compile gates.rego, freeze.rego, risk.rego, etc.
	var policyEngine *policy.OPAEngine
	// Example future shape:
	// policyEngine, err = bootstrap.InitializePolicyEngine(ctx, dbComponents)
	// if err != nil {
	//     log.Fatalf("failed to initialize policy engine: %v", err)
	// }

	// Initialize Authority Console (introspection, not governance imposition)
	console, err := bootstrap.InitializeAuthorityConsole(
		dbComponents.Database,
		dbComponents.Registry,
		dbComponents.Ledger,
	)
	if err != nil {
		log.Fatalf("failed to initialize authority console: %v", err)
	}

	// Start daemon service with graceful shutdown
	if err := service.StartDaemon(config, dbComponents.Database, policyEngine, console); err != nil {
		log.Fatalf("daemon error: %v", err)
	}
}
