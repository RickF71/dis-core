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
//
// This phase performs:
//  1. Config + DB setup
//  2. Schema & ledger initialization
//  3. Bootstrap table and YAML imports
//  4. Policy engine setup
//  5. API + MirrorSpin startup
//
// No canon logic is loaded here — only mutable bootstrap state.
func Run() error {
	// ============================================================
	// 1. CONFIGURATION
	// ============================================================
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Printf("⚠️  No config.yaml found, using defaults: %v", err)
		cfg = &config.Config{}
	}

	// ============================================================
	// 2. DATABASE CONNECTION
	// ============================================================
	database, err := db.Connect(cfg)
	if err != nil {
		return err
	}
	defer database.Close()
	log.Println("✅ Connected to PostgreSQL ledger")

	// ============================================================
	// 3. SCHEMA REGISTRY
	// ============================================================
	reg := schema.NewRegistry()

	// Load core and domain schemas
	if err := reg.LoadDir("./disyaml/schemas"); err != nil {
		log.Printf("⚠️  Core schema load failed: %v", err)
	}
	if err := reg.LoadDir("./disyaml/domains"); err != nil {
		log.Printf("⚠️  Domain schema load failed: %v", err)
	}
	log.Printf("📘 Loaded %d schemas into registry", len(reg.ByKey()))

	// ============================================================
	// 4. LEDGER INITIALIZATION
	// ============================================================
	led, err := ledger.Open(cfg.DatabaseDSN, database, reg)
	if err != nil {
		return err
	}
	defer led.Close()
	log.Println("✅ Ledger ready")

	// Preload domains into the ledger memory view
	domainDir := filepath.Join(".", "disyaml/domains")
	if err := led.BootstrapDomains(reg, domainDir); err != nil {
		log.Printf("⚠️  Domain bootstrap failed: %v", err)
	} else {
		log.Println("✅ Domains loaded into ledger")
	}

	// ============================================================
	// 5. BOOTSTRAP PHASE
	// ============================================================
	log.Println("🚀 Starting bootstrap phase...")

	// 5.1 Ensure all tables exist
	if err := bootstrap.BootstrapAllTables(database); err != nil {
		return fmt.Errorf("bootstrap tables: %w", err)
	}

	// 5.2 Import YAMLs into bootstrap table
	if err := bootstrap.ImportYAML(".", database); err != nil {
		log.Printf("⚠️  Bootstrap import failed: %v", err)
	} else {
		log.Println("✅ Bootstrap YAML import complete")
	}

	log.Println("🎯 Bootstrap phase complete.")

	// ============================================================
	// 6. POLICY ENGINE
	// ============================================================
	base := "./policies"
	opaEngine, err := policy.NewOPAEngine()
	if err != nil {
		return fmt.Errorf("failed to start policy engine: %w", err)
	}
	engine := policy.NewPolicyEngineImpl(opaEngine)
	log.Printf("✅ Policy engine initialized (using %s)", base)

	// ============================================================
	// 7. API SERVER
	// ============================================================
	server := api.NewServer(cfg, led, database)
	server.RegisterEvalRoute(engine)
	log.Println("✅ Registered route(s)")

	// Start MirrorSpin diagnostics
	if err := mirrorspin.Start(database); err != nil {
		return err
	}

	// ============================================================
	// 8. START HTTP SERVER
	// ============================================================
	addr := fmt.Sprintf("%s:%d", cfg.APIHost, cfg.APIPort)
	log.Printf("🚀 DIS-Core v%s starting on %s", cfg.Version, addr)

	return http.ListenAndServe(addr, api.WithCORS(server.Mux()))
}
