package bootstrap

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// BootstrapAllTables ensures all core and subsystem tables exist in dependency order.
func BootstrapAllTables(dbConn *pgx.Conn) error {
	ctx := context.Background()
	fmt.Println("🚀 Bootstrapping all DIS-Core tables...")

	// For this refactor, we'll use direct SQL statements rather than
	// updating all the individual EnsureXTable functions
	tables := []string{
		`CREATE TABLE IF NOT EXISTS domains (
			id TEXT PRIMARY KEY,
			type TEXT,
			version TEXT,
			content JSONB,
			source_file TEXT,
			hash TEXT,
			imported_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS schemas (
			id TEXT PRIMARY KEY,
			version TEXT,
			content JSONB,
			hash TEXT,
			imported_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS overlays (
			id TEXT PRIMARY KEY,
			content JSONB,
			imported_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS policies (
			id TEXT PRIMARY KEY,
			content JSONB,
			imported_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS mirror_events (
			id TEXT PRIMARY KEY,
			event_type TEXT,
			payload JSONB,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS peers (
			id TEXT PRIMARY KEY,
			address TEXT,
			last_seen TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS identities (
			id TEXT PRIMARY KEY,
			public_key TEXT,
			metadata JSONB,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS handshakes (
			id TEXT PRIMARY KEY,
			peer_id TEXT,
			status TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS receipts (
			id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
			type TEXT NOT NULL,
			actor TEXT,
			target TEXT,
			domain TEXT,
			payload JSONB,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS import_warnings (
			id TEXT PRIMARY KEY,
			message TEXT,
			source_file TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
	}

	for i, createSQL := range tables {
		if _, err := dbConn.Exec(ctx, createSQL); err != nil {
			return fmt.Errorf("creating table %d: %w", i, err)
		}
	}

	fmt.Println("✅ All tables ensured.")
	return nil
}
