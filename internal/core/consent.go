package core

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"dis-core/internal/config"
	"dis-core/internal/policy"
	"dis-core/internal/util"
)

// PerformConsentAction validates policy and inserts a receipt.
// Returns: receiptID, nonce, timestamp, signature, error
func PerformConsentAction(db *sql.DB, by string, scope string, providedNonce string, cfg *config.Config, pol *policy.Policy, polSum string) (int64, string, string, string, error) {
	var id string
	if err := db.QueryRow("SELECT id FROM identities ORDER BY created_at DESC LIMIT 1").Scan(&id); err != nil {
		return 0, "", "", "", fmt.Errorf("no identity found, create one first")
	}

	// Policy checks
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
	ts := time.Now().UTC().Format(time.RFC3339Nano)

	// Signature includes policy checksum
	sig := util.Sign(action, id, by, scope, nonce, ts, polSum)

	res, err := db.Exec(`INSERT INTO receipts (identity_id, action, by_domain, scope, nonce, timestamp, policy_checksum, signature)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, action, by, scope, nonce, ts, polSum, sig)
	if err != nil {
		return 0, "", "", "", err
	}

	recID, _ := res.LastInsertId()
	return recID, nonce, ts, sig, nil
}
