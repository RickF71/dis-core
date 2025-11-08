package app

import (
	"context"
	"dis-core/internal/api"
	"dis-core/internal/api/server"
	"dis-core/internal/bootstrap"
	"dis-core/internal/config"
	"dis-core/internal/db"
	"dis-core/internal/ledger"
	"dis-core/internal/schema"
	"fmt"
	"log"
	"net/http"
	"os"
)

func Run() error {
	ctx := context.Background()

	// ============================================================
	// 1. CONFIGURATION (env / defaults only)
	// ============================================================
	host := getenvDefault("DIS_API_HOST", "0.0.0.0")
	port := getenvDefault("DIS_API_PORT", "8080")
	version := getenvDefault("DIS_VERSION", "v1.0-dev")
	dsn := config.DatabaseURL()

	log.Printf("🔧 Using database DSN: %s", dsn)

	// ============================================================
	// 2. DATABASE CONNECTION
	// ============================================================
	database, err := db.ConnectPostgres(dsn)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer database.Close(ctx)
	log.Println("✅ Connected to PostgreSQL ledger")

	// ============================================================
	// 3. SCHEMA REGISTRY
	// ============================================================
	schemaReg := schema.NewRegistry()
	log.Println("📘 Schema registry initialized (DB-native mode)")

	// ============================================================
	// 4. LEDGER INITIALIZATION
	// ============================================================
	led, err := ledger.Open(ctx, dsn, database, schemaReg)
	if err != nil {
		return fmt.Errorf("open ledger: %w", err)
	}
	defer led.Close(ctx)
	log.Println("✅ Ledger ready (DB-native)")

	// ============================================================
	// 5. BOOTSTRAP PHASE (migrations only)
	// ============================================================
	log.Println("🚀 Ensuring tables and baseline state...")
	if err := bootstrap.BootstrapAllTables(database); err != nil {
		return fmt.Errorf("bootstrap tables: %w", err)
	}
	log.Println("🎯 Bootstrap phase complete (no YAML import).")

	// ============================================================
	// 6. CORE INITIALIZATION (schema + null policies)
	// ============================================================
	log.Println("🧱 Initializing DIS-Core canonical data (schema + policies)...")
	if err := BootstrapAuthority(database, schemaReg); err != nil {
		return fmt.Errorf("bootstrap core: %w", err)
	}
	log.Println("✅ Core schema + null policies initialized.")

	// ============================================================
	// 7. API SERVER
	// ============================================================
	addr := fmt.Sprintf("%s:%s", host, port)
	log.Printf("🚀 DIS-Core %s starting on http://%s", version, addr)

	srv := api.New(database, led) // Use the main api package

	handler := server.WithCORS(srv.Handler()) // Use Handler() method

	// Single ListenAndServe call
	return http.ListenAndServe(addr, handler)

}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
