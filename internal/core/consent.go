package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"dis-core/internal/bridge"
	dbpkg "dis-core/internal/db"
	"dis-core/internal/policy"
	"dis-core/internal/util/crypto"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PerformConsentAction validates policy and inserts a receipt.
// Returns: receiptID, nonce, createdAt, signature, error
func PerformConsentAction(pool *pgxpool.Pool, by string, scope string, providedNonce string, pol *policy.Policy, polSum string) (int64, string, string, string, error) {
	var id string
	if err := pool.QueryRow(context.Background(), "SELECT id FROM identities ORDER BY created_at DESC LIMIT 1").Scan(&id); err != nil {
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

	// Determine nonce length (env or default)
	nonceBytes := 16
	if env := os.Getenv("DIS_NONCE_BYTES"); env != "" {
		if n, err := strconv.Atoi(env); err == nil && n > 0 {
			nonceBytes = n
		}
	}

	nonce := providedNonce
	if nonce == "" {
		var genErr error
		nonce, genErr = crypto.RandomNonce(nonceBytes)
		if genErr != nil {
			return 0, "", "", "", genErr
		}
	}

	ts := time.Now().UTC()

	// Signature includes policy checksum
	sig := crypto.Sign(action, id, by, scope, nonce, bridge.CanonicalTime(ts), polSum)

	// Construct receipt record
	// Build standardized panels for consent receipt
	actionPanel := map[string]any{
		"type":  "consent.grant.v1",
		"scope": scope,
	}
	domainPanel := map[string]any{
		"origin_id": by,
	}

	r := &dbpkg.Receipt{
		ID:               fmt.Sprintf("rcpt-%s", nonce[:8]),
		Type:             "bridge-receipt-template.v0",
		Actor:            by,
		Domain:           by,
		OriginDomainID:   by,
		OriginDomainName: "",
		Payload:          map[string]any{"content": fmt.Sprintf("Consent granted by %s for scope '%s'. Sig=%s", by, scope, sig[:16])},
		ActionPanel:      actionPanel,
		DomainPanel:      domainPanel,
		CreatedAt:        ts,
	}

	err := dbpkg.SaveReceipt(context.Background(), pool, r)
	if err != nil {
		return 0, "", "", "", err
	}

	log.Printf("✅ Consent action recorded: by=%s scope=%s receipt_id=%s", by, scope, r.ID)
	return 1, nonce, bridge.CanonicalTime(ts), sig, nil
}
