package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

// EnsureValueReceiptsSchema creates the value_receipts table if missing.
func EnsureValueReceiptsSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS value_receipts (
		id TEXT PRIMARY KEY,
		by TEXT NOT NULL,
		action_ref TEXT NOT NULL,
		substrate_ref TEXT,
		coherence_delta REAL NOT NULL,
		value_vector TEXT,
		observer_field TEXT,
		notes TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_value_by_time ON value_receipts(by, timestamp);
	CREATE INDEX IF NOT EXISTS idx_value_observer ON value_receipts(observer_field);
	`
	_, err := db.Exec(schema)
	return err
}
