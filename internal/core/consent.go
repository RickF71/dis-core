package core

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"dis-core/internal/bridge"
	"dis-core/internal/config"
	dbpkg "dis-core/internal/db"
	"dis-core/internal/policy"
	"dis-core/internal/util"
)

// PerformConsentAction validates policy and inserts a receipt.
// Returns: receiptID, nonce, timestamp, signature, error
func PerformConsentAction(sqlDB *sql.DB, by string, scope string, providedNonce string, cfg *config.Config, pol *policy.Policy, polSum string) (int64, string, string, string, error) {
	var id string
	if err := sqlDB.QueryRow("SELECT id FROM identities ORDER BY created_at DESC LIMIT 1").Scan(&id); err != nil {
		return 0, "", "", "", fmt.Errorf("no identity found, create one first")
	}

	// --- Policy checks ---
	if pol.IsDomainDenied(by) {
		return 0, "", "", "", errors.New("deny:domain.denied")
	}
	if !pol.IsAllowed(by, scope) {
		return 0, "", "", "", errors.New("deny:scope.invalid")
	}

	action := "consent:grant"
	nonce := providedNonce
	if nonce == "" {
		var genErr error
		nonce, genErr = util.RandomNonce(cfg.NonceBytes)
		if genErr != nil {
			return 0, "", "", "", genErr
		}
	}

	ts := time.Now().UTC()

	// Signature includes policy checksum
	sig := util.Sign(action, id, by, scope, nonce, bridge.CanonicalTime(ts), polSum)

	// 1️⃣ Construct new-style receipt record
	r := &dbpkg.Receipt{
		ReceiptID: fmt.Sprintf("rcpt-%s", nonce[:8]),
		SchemaRef: "bridge-receipt-template.v0",
		Content:   fmt.Sprintf("Consent granted by %s for scope '%s'. Sig=%s", by, scope, sig[:16]),
		Timestamp: ts,
	}

	recID, err := dbpkg.InsertReceipt(sqlDB, r)
	if err != nil {
		return 0, "", "", "", err
	}

	log.Printf("✅ Consent action recorded: by=%s scope=%s receipt_id=%d", by, scope, recID)
	return recID, nonce, bridge.CanonicalTime(ts), sig, nil
}
