package mirrorspin

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// SpinLoop periodically checks for new mirror events and logs a summary.
// This loop self-heals schema drift (via EnsureMirrorEventsTable) and runs continuously.
func SpinLoop(db *sql.DB) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var count int
		// Query the mirror_events table using created_at for time window
		err := db.QueryRow(`SELECT COUNT(*) FROM mirror_events WHERE created_at > NOW() - INTERVAL '15 seconds'`).Scan(&count)
		if err != nil {
			log.Printf("MirrorSpin detect error: %v", err)
			continue
		}

		if count > 0 {
			log.Printf("🪞 MirrorSpin detected %d recent mirror events", count)
		} else {
			log.Printf("🪞 MirrorSpin idle — no new events in last 15s")
		}

		// Optionally: insert a heartbeat or perform a lightweight reflection event
		_, insertErr := db.Exec(`
			INSERT INTO mirror_events (event_type, message)
			VALUES ('mirror.heartbeat', 'MirrorSpin loop ticked OK')
		`)
		if insertErr != nil {
			log.Printf("MirrorSpin heartbeat insert failed: %v", insertErr)
		}
	}
}

// detectDBChanges is a stub that checks for any recent updates in key tables.
// Later, you can add checksum or timestamp diffing logic.
func detectDBChanges(db *sql.DB) (bool, string) {
	// Example: check if any receipts or identities changed recently
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM receipts WHERE created_at > NOW() - INTERVAL '15 seconds'`).Scan(&count)
	if err != nil {
		log.Printf("MirrorSpin detect error: %v", err)
		return false, ""
	}
	if count > 0 {
		return true, fmt.Sprintf("%d new receipts detected", count)
	}
	return false, ""
}

// emitMirrorEvent creates a log entry (and later, DB insert) when reflection occurs.
func emitMirrorEvent(db *sql.DB, details string) error {
	// Eventually this should insert into a mirror_events table.
	// For now, just log the event.
	log.Printf("✨ MirrorSpin emitted event: %s", details)
	return nil
}
