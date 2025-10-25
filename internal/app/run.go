package app

import (
	"dis-core/internal/api"
	"dis-core/internal/canon"
	"dis-core/internal/config"
	"dis-core/internal/db"
	"dis-core/internal/ledger"
	"dis-core/internal/mirrorspin"
	"dis-core/internal/policy"
	"fmt"
	"log"
	"net/http"
)

func Run() error {
	// Load config from file, fallback to defaults if missing
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Printf("⚠️  No config.yaml found, using defaults: %v", err)
		cfg = &config.Config{}
	}

	// // Setup context with graceful shutdown
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// go func() {
	// 	c := make(chan os.Signal, 1)
	// 	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// 	<-c
	// 	log.Println("shutting down...")
	// 	cancel()
	// }()

	// Connect to DB and bootstrap schema
	database, err := db.Connect(cfg)
	if err != nil {
		return err
	}
	defer database.Close()
	log.Println("✅ Connected to PostgreSQL ledger")

	if err := db.EnsureTables(database, cfg); err != nil {
		return err
	}
	log.Println("📜 Registered schema(s)")

	// Open ledger
	led, err := ledger.Open(cfg.DatabaseDSN, database)
	if err != nil {
		return err
	}
	defer led.Close()
	log.Println("✅ Ledger ready.")

	// Canon import/freeze/register logic
	if err := canon.Import(database); err != nil {
		return err
	}
	if err := canon.Export(database); err != nil {
		return err
	}
	if err := canon.Freeze(database); err != nil {
		return err
	}

	// Start policy engine
	base := "./policies"
	opaEngine, err := policy.NewOPAEngine()
	if err != nil {
		return fmt.Errorf("failed to start policy engine: %w", err)
	}
	engine := policy.NewPolicyEngineImpl(opaEngine)
	log.Printf("✅ Policy engine initialized (using %s)", base)

	// Start API server
	server := api.NewServer(cfg, led, database)
	server.RegisterEvalRoute(engine) // wire policy engine
	log.Println("✅ Registered route(s)")

	if err := mirrorspin.Start(database); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", cfg.APIHost, cfg.APIPort)
	log.Printf("🚀 DIS-Core v%s starting on %s", cfg.Version, addr)
	return http.ListenAndServe(addr, api.WithCORS(server.Mux()))
}
