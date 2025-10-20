package db

import (
	"database/sql"
)

// EnsureLifePushSubstrateSchema creates the lifepush_substrate_structure table (PostgreSQL version).
func EnsureLifePushSubstrateSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS lifepush_substrate_structure (
		id TEXT PRIMARY KEY,
		layer TEXT NOT NULL,
		coherence_threshold DOUBLE PRECISION NOT NULL,
		energy_flow_min DOUBLE PRECISION NOT NULL,
		consent_integrity_min DOUBLE PRECISION NOT NULL,
		successor_layer TEXT,
		observer_domain TEXT,
		active BOOLEAN DEFAULT TRUE
	);

	CREATE INDEX IF NOT EXISTS idx_substrate_layer 
		ON lifepush_substrate_structure(layer);
	`
	_, err := db.Exec(schema)
	return err
}
