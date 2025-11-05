package app

import (
	"dis-core/internal/api"
	"dis-core/internal/bootstrap"
	"dis-core/internal/canon"
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
	dsn := cfg.DatabaseURL()
	database, err := db.ConnectDSN(dsn)
	if err != nil {
		return err
	}
	defer database.Close()
	log.Println("✅ Connected to PostgreSQL ledger")

	// ============================================================
	// 3. SCHEMA REGISTRY
	// ============================================================
	schemaReg := schema.NewRegistry()

	if err := schemaReg.LoadDir("./disyaml/schemas"); err != nil {
		log.Printf("⚠️  Core schema load failed: %v", err)
	}
	if err := schemaReg.LoadDir("./disyaml/domains"); err != nil {
		log.Printf("⚠️  Domain schema load failed: %v", err)
	}
	log.Printf("📘 Loaded %d schemas into registry", len(schemaReg.ByKey()))

	// ============================================================
	// 3.5 CANON REGISTRY AND THEME BOOTSTRAP
	// ============================================================
	canonReg := &canon.Registry{DB: database}

	log.Println("🌍 Bootstrapping canonical domain themes...")
	canon.BootstrapThemes(canonReg)
	log.Println("✅ Canonical domain themes seeded (Terra + USA).")

	// ============================================================
	// 4. LEDGER INITIALIZATION
	// ============================================================
	led, err := ledger.Open(cfg.DatabaseDSN, database, schemaReg)
	if err != nil {
		return err
	}
	defer led.Close()
	log.Println("✅ Ledger ready")

	// Preload domains into the ledger memory view
	domainDir := filepath.Join(".", "disyaml/domains")
	if err := led.BootstrapDomains(schemaReg, domainDir); err != nil {
		log.Printf("⚠️  Domain bootstrap failed: %v", err)
	} else {
		log.Println("✅ Domains loaded into ledger")
	}

	// ============================================================
	// 5. BOOTSTRAP PHASE
	// ============================================================
	log.Println("🚀 Starting bootstrap phase...")

	if err := bootstrap.BootstrapAllTables(database); err != nil {
		return fmt.Errorf("bootstrap tables: %w", err)
	}
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
