package receipts

import (
	"context"
	"time"

	"dis-core/internal/contextx"
	"dis-core/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// EmitRootPSeatClaimTx creates and saves a domain.null root pseat claim receipt
// inside the provided transaction. It uses the canonical SaveReceiptTx path so
// the insertion is panel-aware and atomic with surrounding updates.
func EmitRootPSeatClaimTx(ctx context.Context, tx pgx.Tx, domainID, actorID, pseatID, nonce string) (string, error) {
	if tx == nil {
		return "", ErrNoTx
	}

	id := uuid.NewString()
	now := time.Now().UTC()

	// Build payload
	payload := map[string]any{
		"domain_id": domainID,
		"actor_id":  actorID,
		"pseat_id":  pseatID,
		"nonce":     nonce,
	}

	// Panels
	actionPanel := map[string]any{"type": "root_pseat_claim", "pseat_id": pseatID}
	identityPanel := map[string]any{"identity_id": actorID}
	domainPanel := map[string]any{"domain_id": "null"}

	var policyPanel map[string]any
	if dm, ok := contextx.PolicyDecisionMapFromContext(ctx); ok && dm != nil {
		cp := map[string]any{}
		for k, v := range dm {
			cp[k] = v
		}
		policyPanel = cp
	}

	r := &db.Receipt{
		ID:               id,
		Type:             "domain.null.pseat.claim.v1",
		Actor:            actorID,
		Target:           pseatID,
		Domain:           domainID,
		Payload:          payload,
		CreatedAt:        now,
		OriginDomainID:   domainID,
		OriginDomainName: "null",
		ActionPanel:      actionPanel,
		PolicyPanel:      policyPanel,
		IdentityPanel:    identityPanel,
		DomainPanel:      domainPanel,
	}

	if err := db.SaveReceiptTx(ctx, tx, r); err != nil {
		return "", err
	}
	return id, nil
}

// ErrNoTx is returned when a tx-aware emitter is called without a transaction.
var ErrNoTx = pgx.ErrTxClosed
