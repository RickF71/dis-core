package phase9c

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ListReceipts returns a paginated list of receipts
func ListReceipts(ctx context.Context, pool *pgxpool.Pool, limit, offset int) ([]Receipt, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}
	if limit > 1000 {
		limit = 1000 // Maximum limit
	}

	query := `
		SELECT id, receipt_type, event_id,
		       COALESCE(policy_ref, '') as policy_ref,
		       COALESCE(redaction_ref, '') as redaction_ref,
		       COALESCE(issued_by, '') as issued_by,
		       issued_at, verified,
		       COALESCE(metadata, '{}'::jsonb) as metadata
		FROM receipts
		ORDER BY issued_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query receipts: %w", err)
	}
	defer rows.Close()

	var receipts []Receipt
	for rows.Next() {
		var r Receipt
		if err := rows.Scan(&r.ID, &r.ReceiptType, &r.EventID, &r.PolicyRef, &r.RedactionRef, &r.IssuedBy, &r.IssuedAt, &r.Verified, &r.Metadata); err != nil {
			return nil, fmt.Errorf("failed to scan receipt: %w", err)
		}
		receipts = append(receipts, r)
	}

	return receipts, nil
}

// GetReceiptStats returns statistics about receipts for the dashboard
func GetReceiptStats(ctx context.Context, pool *pgxpool.Pool) (ReceiptStats, error) {
	var stats ReceiptStats

	// Count total receipts
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM receipts").Scan(&stats.Total)
	if err != nil {
		return stats, fmt.Errorf("failed to count total receipts: %w", err)
	}

	// Count verified receipts
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM receipts WHERE verified = true").Scan(&stats.Verified)
	if err != nil {
		return stats, fmt.Errorf("failed to count verified receipts: %w", err)
	}

	// Count orphan receipts (missing policy_ref OR redaction_ref)
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM receipts WHERE policy_ref IS NULL OR redaction_ref IS NULL").Scan(&stats.Orphans)
	if err != nil {
		return stats, fmt.Errorf("failed to count orphan receipts: %w", err)
	}

	// Count by receipt type
	rows, err := pool.Query(ctx, "SELECT receipt_type, COUNT(*) FROM receipts GROUP BY receipt_type")
	if err != nil {
		return stats, fmt.Errorf("failed to count by type: %w", err)
	}
	defer rows.Close()

	stats.ByType = make(map[string]int)
	for rows.Next() {
		var receiptType string
		var count int
		if err := rows.Scan(&receiptType, &count); err != nil {
			return stats, fmt.Errorf("failed to scan type count: %w", err)
		}
		stats.ByType[receiptType] = count
	}

	stats.Timestamp = time.Now().UTC().Format(time.RFC3339)
	return stats, nil
}

// VerifyReceipt verifies a Phase 9C receipt by ID and returns verification result
func VerifyReceipt(ctx context.Context, pool *pgxpool.Pool, id string) (VerificationResult, error) {
	var r Receipt
	var result VerificationResult

	err := pool.QueryRow(ctx, `
		SELECT id, receipt_type, event_id,
		       COALESCE(policy_ref, '') as policy_ref,
		       COALESCE(redaction_ref, '') as redaction_ref,
		       issued_by, issued_at, verified
		FROM receipts WHERE id = $1
	`, id).Scan(&r.ID, &r.ReceiptType, &r.EventID, &r.PolicyRef, &r.RedactionRef, &r.IssuedBy, &r.IssuedAt, &r.Verified)

	if err != nil {
		return result, fmt.Errorf("receipt not found: %w", err)
	}

	result.ReceiptID = r.ID
	result.PolicyRef = r.PolicyRef
	result.RedactionRef = r.RedactionRef
	result.Timestamp = time.Now().UTC().Format(time.RFC3339)

	// Check for issues
	if r.PolicyRef == "" {
		result.Issues = append(result.Issues, "missing policy_ref")
	}
	if r.RedactionRef == "" {
		result.Issues = append(result.Issues, "missing redaction_ref")
	}

	result.Verified = len(result.Issues) == 0
	return result, nil
}
