package daemon

import (
	"context"
	"log"
	"time"

	"dis-core/internal/db"
)

// AutoRevocationDaemon scans for expired, non-revoked handshakes
// and emits revocation receipts automatically.
//
// Integration points expected (once db layer is ready):
//   - db.ListExpiredActiveHandshakes(now time.Time) ([]db.Handshake, error)
//   - db.MarkHandshakeRevoked(id int64, when time.Time, reason string) error
//   - db.SaveReceipt(r db.Receipt) error
//
// Until then, stubbed safe no-op versions are provided below.
func StartAutoRevocationDaemon(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	log.Printf("⚙️  Auto-Revocation Daemon starting (interval=%s)", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	run := func() {
		now := time.Now().UTC()

		list, err := safeListExpiredActiveHandshakes(now)
		if err != nil {
			log.Printf("auto-revoke: list error: %v", err)
			return
		}
		if len(list) == 0 {
			return
		}

		for _, hs := range list {
			reason := "expired"
			if err := safeMarkHandshakeRevoked(hs.ID, now, reason); err != nil {
				log.Printf("auto-revoke: mark revoked failed (id=%d token=%s): %v", hs.ID, hs.Token, err)
				continue
			}

			rc := db.Receipt{
				ReceiptID: generateReceiptID("rcpt-revoke"),
				SchemaRef: "revocation.v0",
				Content:   "Revocation: handshake " + hs.Token + " for " + hs.Subject + " (reason: expired)",
				CreatedAt: now, // now is time.Time
			}

			if err := safeSaveReceipt(rc); err != nil {
				log.Printf("auto-revoke: save receipt failed for token=%s: %v", hs.Token, err)
				continue
			}
			log.Printf("auto-revoke: handshake %s revoked + receipt emitted", hs.Token)
		}
	}

	// initial pass shortly after boot
	go func() {
		time.Sleep(2 * time.Second)
		run()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("⚙️  Auto-Revocation Daemon stopping")
			return
		case <-ticker.C:
			run()
		}
	}
}

func generateReceiptID(prefix string) string {
	// Keep aligned with your existing style: rcpt-<kind>-YYYYMMDDHHMMSS
	return prefix + "-" + time.Now().UTC().Format("20060102150405")
}

//
// ---- Temporary stubs to allow compilation ----
//

// Handshake stub structure (remove when db.Handshake exists)
type handshakeStub struct {
	ID      int64
	Token   string
	Subject string
}

// safeListExpiredActiveHandshakes is a temporary wrapper to prevent compile errors.
func safeListExpiredActiveHandshakes(now time.Time) ([]handshakeStub, error) {
	// Replace with: return db.ListExpiredActiveHandshakes(now)
	return []handshakeStub{}, nil
}

// safeMarkHandshakeRevoked is a placeholder for db.MarkHandshakeRevoked.
func safeMarkHandshakeRevoked(id int64, when time.Time, reason string) error {
	// Replace with: return db.MarkHandshakeRevoked(id, when, reason)
	return nil
}

// safeSaveReceipt safely wraps db.SaveReceipt.
func safeSaveReceipt(r db.Receipt) error {
	// Replace with: return db.SaveReceipt(r)
	return nil
}
