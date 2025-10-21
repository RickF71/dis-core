package net

import "database/sql"

// EnsurePeersTable creates the peers table if it doesn't exist.
func EnsurePeersTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS peers (
		id TEXT PRIMARY KEY,
		address TEXT NOT NULL,
		last_seen TIMESTAMPTZ DEFAULT NOW(),
		status TEXT DEFAULT 'unknown'
	);
	`)
	return err
}
