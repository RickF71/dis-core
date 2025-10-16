package daemon

import (
	"context"
	"dis-core/internal/db"

	"log"
	"time"
)

// AutoRevocationDaemon scans for expired, non-revoked handshakes and
// emits revocation receipts automatically.
//
// Integration expectations in internal/db (keep these slim):
//   - ListExpiredActiveHandshakes(now time.Time) ([]db.Handshake, error)
//   - MarkHandshakeRevoked(id int64, when time.Time, reason string) error
//   - SaveReceipt(r db.Receipt) error
//
// Receipt guidance:
//
//	r.SchemaRef: "revocation.v0" (or your preferred template name)
//	r.Content:   e.g. "Revocation: handshake <token> for <subject> (reason: expired)"
func StartAutoRevocationDaemon(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	log.Printf("⚙️  Auto-Revocation Daemon starting (interval=%s)", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	run := func() {
		now := time.Now().UTC()
		list, err := db.ListExpiredActiveHandshakes(now)
		if err != nil {
			log.Printf("auto-revoke: list error: %v", err)
			return
		}
		if len(list) == 0 {
			return
		}
		for _, hs := range list {
			reason := "expired"
			if err := db.MarkHandshakeRevoked(hs.ID, now, reason); err != nil {
				log.Printf("auto-revoke: mark revoked failed (id=%d token=%s): %v", hs.ID, hs.Token, err)
				continue
			}
			// Emit a revocation receipt
			rc := db.Receipt{
				ReceiptID: generateReceiptID("rcpt-revoke"),
				SchemaRef: "revocation.v0",
				Content:   "Revocation: handshake " + hs.Token + " for " + hs.Subject + " (reason: expired)",
				Timestamp: now.Format(time.RFC3339Nano),
			}
			if err := db.SaveReceipt(rc); err != nil {
				log.Printf("auto-revoke: save receipt failed for token=%s: %v", hs.Token, err)
				// best-effort; continue
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
