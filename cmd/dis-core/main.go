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

	// Initialize database components
	dbComponents, err := bootstrap.InitializeDatabase(ctx, config.DSN)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer dbComponents.Close()

	// Bootstrap database tables
	if err := bootstrap.BootstrapTables(dbComponents.Database); err != nil {
		log.Fatalf("failed to bootstrap tables: %v", err)
	}

	// Initialize Authority Console
	console, err := bootstrap.InitializeAuthorityConsole(dbComponents.Database, dbComponents.Registry, dbComponents.Ledger)
	if err != nil {
		log.Fatalf("failed to initialize authority console: %v", err)
	}

	// Initialize policy engine (optional)
	var policyEngine *policy.OPAEngine
	// Note: Policy engine initialization can be added here if needed

	// Start daemon service with graceful shutdown
	if err := service.StartDaemon(config, dbComponents.Database, policyEngine, console); err != nil {
		log.Fatalf("daemon error: %v", err)
	}
}
