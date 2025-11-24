package main

import (
	"context"
	"log"

	"dis-core/cmd/dis-core/bootstrap"
	"dis-core/cmd/dis-core/service"
	coreauth "dis-core/internal/core/authority"
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

	// 🔵 BedrockBootstrap: one-time, human-approved creation of the 1D root domain `null`
	if err := bootstrap.RunBedrockBootstrap(ctx, dbComponents.Database); err != nil {
		log.Fatalf("bedrock bootstrap failed: %v", err)
	}

	// Initialize policy engine (OPA/Rego)
	var policyEngine *policy.OPAEngine
	if pc, perr := bootstrap.InitializePolicyEngine(); perr == nil && pc != nil {
		if pe, ok := pc.Engine.(*policy.OPAEngine); ok {
			policyEngine = pe
		}
	} else if perr != nil {
		log.Printf("warning: failed to initialize policy engine: %v", perr)
	}

	// Initialize Authority Console (introspection)
	console, err := bootstrap.InitializeAuthorityConsole(
		dbComponents.Database,
		dbComponents.Registry,
		dbComponents.Ledger,
	)
	if err != nil {
		log.Fatalf("failed to initialize authority console: %v", err)
	}

	// Initialize the new core authority engine and pass it into the daemon
	authorityCfg := &coreauth.Config{}
	authorityEngine := coreauth.NewEngine(authorityCfg, dbComponents.Database)
	// Attach the global policy engine to the authority engine so EvaluateTx
	// and other enforcement helpers can access policy at runtime.
	if policyEngine != nil {
		// Wrap the PolicyEngine into the authority engine's EvalFn to avoid
		// import cycles and provide a stable evaluation callback.
		authorityEngine.SetPolicyEvalFunc(func(ctx context.Context, input map[string]interface{}) (bool, string, map[string]interface{}, error) {
			d, err := policyEngine.EvaluateAction(ctx, input)
			if err != nil {
				return false, "", nil, err
			}
			if d == nil {
				return false, "no decision", nil, nil
			}
			return d.Allow, d.Reason, d.Details, nil
		})
	}

	// Start daemon service with graceful shutdown
	if err := service.StartDaemon(config, dbComponents.Database, policyEngine, console, authorityEngine); err != nil {
		log.Fatalf("daemon error: %v", err)
	}
}
