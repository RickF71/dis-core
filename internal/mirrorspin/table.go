package mirrorspin

import (
	"database/sql"
	"log"
)

// EnsureMirrorEventsTable creates the mirror_events table with timestamp, message, hash, and payload.
func EnsureMirrorEventsTable(db *sql.DB) error {
	_, err := db.Exec(`
	       CREATE TABLE IF NOT EXISTS mirror_events (
		       id SERIAL PRIMARY KEY,
		       event_type TEXT NOT NULL DEFAULT 'mirror.spin',
		       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		       message TEXT,
		       hash TEXT,
		       source TEXT,
		       target TEXT,
		       payload JSONB DEFAULT '{}'::jsonb
	       );
       `)
	if err != nil {
		return err
	}
	log.Println("✅ mirror_events table ready")
	return nil
}
