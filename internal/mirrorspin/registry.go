package mirrorspin

import (
	"database/sql"
	"log"
)

// EnsureMirrorEventsTable ensures the mirror_events table exists.
func EnsureMirrorEventsTable(db *sql.DB) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS mirror_events (
		id SERIAL PRIMARY KEY,
		event_type TEXT NOT NULL DEFAULT 'mirror.spin',
		timestamp TIMESTAMPTZ DEFAULT NOW(),
		message TEXT,
		hash TEXT,
		source TEXT,
		target TEXT
	);
	`
	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}
	log.Println("✅ mirror_events table ready")
	return nil
}
