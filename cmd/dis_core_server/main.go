package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"dis-core/internal/api"
	"dis-core/internal/config"
	"dis-core/internal/policy"

	_ "modernc.org/sqlite" // SQLite driver
)

func main() {
	// Allow optional config path via env var or default to config/config.yaml
	configPath := os.Getenv("DIS_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("❌ Failed to load config (%s): %v", configPath, err)
	}

	// Open database
	store, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		log.Fatalf("❌ Failed to open database: %v", err)
	}
	defer store.Close()

	// Initialize policy and core hash placeholders
	pol := &policy.Policy{}
	sum := "development"
	coreHash := "dev"

	// Create server and start
	server := api.NewServer(store, cfg, pol, sum, coreHash)

	addr := cfg.APIHost + ":" + fmt.Sprint(cfg.APIPort)
	log.Printf("🚀 Launching dis_core_server using %s ...", configPath)
	if err := server.Start(addr); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
