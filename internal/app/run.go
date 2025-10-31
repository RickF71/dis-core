package app

import (
	"dis-core/internal/api"
	"dis-core/internal/bootstrap"
	"dis-core/internal/config"
	"dis-core/internal/db"
	"dis-core/internal/ledger"
	"dis-core/internal/mirrorspin"
	"dis-core/internal/policy"
	"dis-core/internal/schema"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
)

// Run initializes and starts the DIS-Core service.
// The bootstrap layer now handles all table creation and YAML imports.
// No canon logic is used here — only editable bootstrap state.
func Run() error {
	// ------------------------------------------------------------
	// 0. Load configuration
	// ------------------------------------------------------------
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Printf("⚠️  No config.yaml found, using defaults: %v", err)
		cfg = &config.Config{}
	}

	// ------------------------------------------------------------
	// 1. Connect to database
	// ------------------------------------------------------------
	database, err := db.Connect(cfg)
	if err != nil {
		return err
	}
	defer database.Close()
	log.Println("✅ Connected to PostgreSQL ledger")

	// ------------------------------------------------------------
	// 2. Initialize schema registry
	// ------------------------------------------------------------
	reg := schema.NewRegistry()

	// Load schemas from disyaml tree
	if err := reg.LoadDir("./disyaml/schemas"); err != nil {
		log.Printf("⚠️  Core schema load failed: %v", err)
	}
	if err := reg.LoadDir("./disyaml/domains"); err != nil {
		log.Printf("⚠️  Domain schema load failed: %v", err)
	}
	log.Printf("📘 Loaded %d schemas into registry", len(reg.ByKey()))

	// ------------------------------------------------------------
	// 3. Open ledger and load domain scaffolds
	// ------------------------------------------------------------
	led, err := ledger.Open(cfg.DatabaseDSN, database, reg)
	if err != nil {
		return err
	}
	defer led.Close()
	log.Println("✅ Ledger ready")

	domainDir := filepath.Join(".", "disyaml/domains")
	if err := led.BootstrapDomains(reg, domainDir); err != nil {
		log.Printf("⚠️  Domain bootstrap failed: %v", err)
	} else {
		log.Println("✅ Domains loaded into ledger")
	}

	// ------------------------------------------------------------
	// 4. Unified Bootstrap Phase
	// ------------------------------------------------------------
	log.Println("🚀 Starting bootstrap phase...")

	// 4.1 Ensure all tables exist
	if err := bootstrap.BootstrapAllTables(database); err != nil {
		return fmt.Errorf("bootstrap tables: %w", err)
	}

	// 4.2 Import all YAML files into the bootstrap layer
	if err := bootstrap.ImportYAML(".", database); err != nil {
		log.Printf("⚠️  Bootstrap import failed: %v", err)
	} else {
		log.Println("✅ Bootstrap YAML import complete")
	}

	log.Println("🎯 Bootstrap phase complete.")

	// ------------------------------------------------------------
	// 5. Initialize policy engine
	// ------------------------------------------------------------
	base := "./policies"
	opaEngine, err := policy.NewOPAEngine()
	if err != nil {
		return fmt.Errorf("failed to start policy engine: %w", err)
	}
	engine := policy.NewPolicyEngineImpl(opaEngine)
	log.Printf("✅ Policy engine initialized (using %s)", base)

	// ------------------------------------------------------------
	// 6. Start API server
	// ------------------------------------------------------------
	server := api.NewServer(cfg, led, database)
	server.RegisterEvalRoute(engine)
	log.Println("✅ Registered route(s)")

	if err := mirrorspin.Start(database); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", cfg.APIHost, cfg.APIPort)
	log.Printf("🚀 DIS-Core v%s starting on %s", cfg.Version, addr)
	return http.ListenAndServe(addr, api.WithCORS(server.Mux()))
}
