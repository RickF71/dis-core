package mirrorspin

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// SpinLoop runs the MirrorSpin engine on an interval, watching for DB deltas.
func SpinLoop(db *sql.DB) {
	tick := time.NewTicker(15 * time.Second)
	defer tick.Stop()

	for range tick.C {
		changed, details := detectDBChanges(db)
		if changed {
			if err := emitMirrorEvent(db, details); err != nil {
				log.Printf("❌ MirrorSpin event failed: %v", err)
			} else {
				log.Printf("🪞 MirrorSpin: %s", details)
			}
		} else {
			log.Println("🪞 MirrorSpin: no changes detected")
		}
	}
}

// detectDBChanges is a stub that checks for any recent updates in key tables.
// Later, you can add checksum or timestamp diffing logic.
func detectDBChanges(db *sql.DB) (bool, string) {
	// Example: check if any receipts or identities changed recently
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM receipts WHERE timestamp > NOW() - INTERVAL '15 seconds'`).Scan(&count)
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
