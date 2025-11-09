package bootstrap

import (
	"context"
	"fmt"
	"log"

	"dis-core/internal/bootstrap"
	"dis-core/internal/db"
	"dis-core/internal/ledger"
	"dis-core/internal/schema"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DatabaseComponents holds the initialized database-related components
type DatabaseComponents struct {
	Database *pgxpool.Pool
	Registry *schema.Registry
	Ledger   *ledger.Ledger
}

// InitializeDatabase sets up the database connection, schema registry, and ledger
func InitializeDatabase(ctx context.Context, dsn string) (*DatabaseComponents, error) {
	log.Printf("🔧 Using database DSN: %s", dsn)

	// Connect to database
	database, err := db.ConnectPostgres(dsn)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	log.Println("✅ Connected to PostgreSQL ledger")

	// Initialize schema registry
	schemaReg := schema.NewRegistry()
	log.Println("📘 Schema registry initialized (DB-native mode)")

	// Initialize ledger
	led, err := ledger.Open(ctx, dsn, database, schemaReg)
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("open ledger: %w", err)
	}
	log.Println("✅ Ledger ready (DB-native)")

	return &DatabaseComponents{
		Database: database,
		Registry: schemaReg,
		Ledger:   led,
	}, nil
}

// BootstrapTables ensures all database tables are created
func BootstrapTables(database *pgxpool.Pool) error {
	log.Println("🚀 Ensuring tables and baseline state...")
	if err := bootstrap.BootstrapAllTables(database); err != nil {
		return fmt.Errorf("bootstrap tables: %w", err)
	}
	log.Println("🎯 Bootstrap phase complete (no YAML import).")
	return nil
}

// Close properly closes all database components
func (dc *DatabaseComponents) Close() {
	if dc.Ledger != nil {
		dc.Ledger.Close()
	}
	if dc.Database != nil {
		dc.Database.Close()
	}
}
