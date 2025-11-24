package receipts

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// VerifyReceipt loads a receipt by id and performs simple continuity checks.
func VerifyReceipt(ctx context.Context, pool *pgxpool.Pool, id string) (VerificationResult, error) {
	var r Receipt
	var result VerificationResult

	err := pool.QueryRow(ctx, `
		SELECT id, receipt_type, event_id, COALESCE(policy_ref, ''), COALESCE(redaction_ref, ''), issued_by, issued_at, verified
		FROM receipts_9c WHERE id = $1
	`, id).Scan(&r.ID, &r.ReceiptType, &r.EventID, &r.PolicyRef, &r.RedactionRef, &r.IssuedBy, &r.IssuedAt, &r.Verified)
	if err != nil {
		return result, fmt.Errorf("receipt not found: %w", err)
	}

	result.ReceiptID = r.ID
	result.PolicyRef = r.PolicyRef
	result.RedactionRef = r.RedactionRef
	result.Timestamp = time.Now().UTC().Format(time.RFC3339)

	if r.PolicyRef == "" {
		result.Issues = append(result.Issues, "missing policy_ref")
	}
	if r.RedactionRef == "" {
		result.Issues = append(result.Issues, "missing redaction_ref")
	}
	result.Verified = len(result.Issues) == 0
	return result, nil
}
