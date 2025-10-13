package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// EnsureLifePushSubstrateSchema creates the lifepush_substrate_structure table.
func EnsureLifePushSubstrateSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS lifepush_substrate_structure (
		id TEXT PRIMARY KEY,
		layer TEXT NOT NULL,
		coherence_threshold REAL NOT NULL,
		energy_flow_min REAL NOT NULL,
		consent_integrity_min REAL NOT NULL,
		successor_layer TEXT,
		observer_domain TEXT,
		active INTEGER DEFAULT 1
	);
	CREATE INDEX IF NOT EXISTS idx_substrate_layer ON lifepush_substrate_structure(layer);
	`
	_, err := db.Exec(schema)
	return err
}
