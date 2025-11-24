package db

import (
	"context"
	"fmt"
	"time"

	"dis-core/internal/contextx"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ClaimRootPSeatTx assigns the null domain prime seat to the provided actorID
// within the provided transaction. It verifies the null domain and prime seat
// exist and are unoccupied, performs a safe UPDATE ... RETURNING, and emits a
// receipt using the unified SaveReceiptTx path.
func ClaimRootPSeatTx(ctx context.Context, tx pgx.Tx, actorID string) error {
	if tx == nil {
		return fmt.Errorf("ClaimRootPSeatTx: tx is nil")
	}

	// Resolve null domain id
	var nullDomainID string
	if err := tx.QueryRow(ctx, `SELECT id FROM domains WHERE name = 'null' OR name = 'domain.null' LIMIT 1`).Scan(&nullDomainID); err != nil {
		return fmt.Errorf("claim root pseat: resolve null domain: %w", err)
	}

	// Find the prime seat for the null domain
	var seatID string
	var memberID *string
	if err := tx.QueryRow(ctx, `SELECT id, member_id FROM domain_seats WHERE domain_id = $1 AND seat_type = 'prime' LIMIT 1`, nullDomainID).Scan(&seatID, &memberID); err != nil {
		return fmt.Errorf("claim root pseat: find prime seat: %w", err)
	}

	if memberID != nil {
		return fmt.Errorf("claim root pseat: already occupied by %s", *memberID)
	}

	// Attempt to safely assign the occupant using UPDATE ... RETURNING to avoid races
	var updatedID string
	if err := tx.QueryRow(ctx, `UPDATE domain_seats SET member_id = $1 WHERE id = $2 AND member_id IS NULL RETURNING id`, actorID, seatID).Scan(&updatedID); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("claim root pseat: already occupied")
		}
		return fmt.Errorf("claim root pseat: update failed: %w", err)
	}

	// Emit receipt for the claim using the canonical SaveReceiptTx entrypoint.
	payload := map[string]any{
		"domain_id": nullDomainID,
		"actor_id":  actorID,
		"pseat_id":  updatedID,
		"nonce":     uuid.NewString(),
	}

	actionPanel := map[string]any{"type": "root_pseat_claim", "pseat_id": updatedID}
	identityPanel := map[string]any{"identity_id": actorID}
	domainPanel := map[string]any{"domain_id": "null"}

	// Attach policy decision map from context so PolicyPanel is populated
	// when SaveReceiptTx writes panels-aware rows.
	var policyPanel map[string]any
	if dm, ok := contextx.PolicyDecisionMapFromContext(ctx); ok && dm != nil {
		cp := map[string]any{}
		for k, v := range dm {
			cp[k] = v
		}
		policyPanel = cp
	}

	r := &Receipt{
		ID:               uuid.NewString(),
		Type:             "domain.null.pseat.claim.v1",
		Actor:            actorID,
		Target:           updatedID,
		Domain:           nullDomainID,
		Payload:          payload,
		CreatedAt:        time.Now().UTC(),
		OriginDomainID:   nullDomainID,
		OriginDomainName: "null",
		ActionPanel:      actionPanel,
		IdentityPanel:    identityPanel,
		DomainPanel:      domainPanel,
		PolicyPanel:      policyPanel,
	}

	if err := SaveReceiptTx(ctx, tx, r); err != nil {
		return fmt.Errorf("claim root pseat: emit receipt: %w", err)
	}

	return nil
}
