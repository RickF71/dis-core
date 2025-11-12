package receipts

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PolicyContinuityResult represents the policy continuity validation for a domain
type PolicyContinuityResult struct {
	DomainRef      string          `json:"domain_ref"`
	Timestamp      string          `json:"timestamp"`
	TotalReceipts  int             `json:"total_receipts"`
	ValidRefs      int             `json:"valid_refs"`
	OrphanRefs     int             `json:"orphan_refs"`
	InvalidRefs    int             `json:"invalid_refs"`
	PolicyMappings []PolicyMapping `json:"policy_mappings"`
	OrphanReceipts []OrphanReceipt `json:"orphan_receipts"`
	ContinuityOK   bool            `json:"continuity_ok"`
}

// PolicyMapping represents a valid policy reference mapping
type PolicyMapping struct {
	PolicyRef    string `json:"policy_ref"`
	ReceiptCount int    `json:"receipt_count"`
	PolicyExists bool   `json:"policy_exists"`
	LastSeen     string `json:"last_seen"`
}

// OrphanReceipt represents a receipt with missing or invalid policy reference
type OrphanReceipt struct {
	ReceiptID   string `json:"receipt_id"`
	EventID     string `json:"event_id"`
	PolicyRef   string `json:"policy_ref"`
	IssuedAt    string `json:"issued_at"`
	IssueReason string `json:"issue_reason"`
}

// VerifyPolicyContinuity validates policy continuity for a specific domain
func VerifyPolicyContinuity(ctx context.Context, pool *pgxpool.Pool, domainID string) (PolicyContinuityResult, error) {
	result := PolicyContinuityResult{
		DomainRef:      domainID,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		PolicyMappings: []PolicyMapping{},
		OrphanReceipts: []OrphanReceipt{},
	}

	// Step 1: Get all receipts that should be linked to this domain
	// For now, we'll look for receipts where event_id contains the domain or policy_ref references domain policies
	receiptsQuery := `
		SELECT id, receipt_type, event_id,
		       COALESCE(policy_ref, '') as policy_ref,
		       COALESCE(redaction_ref, '') as redaction_ref,
		       issued_at
		FROM receipts
		WHERE event_id ILIKE $1 OR policy_ref ILIKE $2
		ORDER BY issued_at DESC
	`

	domainPattern := "%" + domainID + "%"
	rows, err := pool.Query(ctx, receiptsQuery, domainPattern, domainPattern)
	if err != nil {
		return result, fmt.Errorf("failed to query domain receipts: %w", err)
	}
	defer rows.Close()

	var receipts []struct {
		ID           string
		ReceiptType  string
		EventID      string
		PolicyRef    string
		RedactionRef string
		IssuedAt     time.Time
	}

	for rows.Next() {
		var r struct {
			ID           string
			ReceiptType  string
			EventID      string
			PolicyRef    string
			RedactionRef string
			IssuedAt     time.Time
		}
		if err := rows.Scan(&r.ID, &r.ReceiptType, &r.EventID, &r.PolicyRef, &r.RedactionRef, &r.IssuedAt); err != nil {
			return result, fmt.Errorf("failed to scan receipt: %w", err)
		}
		receipts = append(receipts, r)
	}

	result.TotalReceipts = len(receipts)

	// Step 2: Analyze policy references
	policyRefCounts := make(map[string]int)
	policyLastSeen := make(map[string]time.Time)

	for _, receipt := range receipts {
		if receipt.PolicyRef == "" {
			// Orphan receipt - no policy reference
			result.OrphanRefs++
			result.OrphanReceipts = append(result.OrphanReceipts, OrphanReceipt{
				ReceiptID:   receipt.ID,
				EventID:     receipt.EventID,
				PolicyRef:   receipt.PolicyRef,
				IssuedAt:    receipt.IssuedAt.Format(time.RFC3339),
				IssueReason: "missing policy_ref",
			})
		} else {
			// Has policy reference - count it
			policyRefCounts[receipt.PolicyRef]++
			if receipt.IssuedAt.After(policyLastSeen[receipt.PolicyRef]) {
				policyLastSeen[receipt.PolicyRef] = receipt.IssuedAt
			}
		}
	}

	// Step 3: Validate policy references against policy registry
	for policyRef, count := range policyRefCounts {
		policyExists, err := validatePolicyExists(ctx, pool, policyRef)
		if err != nil {
			// Consider it invalid if we can't validate
			result.InvalidRefs++
			result.OrphanReceipts = append(result.OrphanReceipts, OrphanReceipt{
				ReceiptID:   "multiple", // This represents multiple receipts
				EventID:     "policy-validation",
				PolicyRef:   policyRef,
				IssuedAt:    policyLastSeen[policyRef].Format(time.RFC3339),
				IssueReason: fmt.Sprintf("policy validation failed: %v", err),
			})
		} else if policyExists {
			result.ValidRefs += count
		} else {
			result.InvalidRefs += count
			result.OrphanReceipts = append(result.OrphanReceipts, OrphanReceipt{
				ReceiptID:   "multiple",
				EventID:     "policy-validation",
				PolicyRef:   policyRef,
				IssuedAt:    policyLastSeen[policyRef].Format(time.RFC3339),
				IssueReason: "policy_ref not found in registry",
			})
		}

		result.PolicyMappings = append(result.PolicyMappings, PolicyMapping{
			PolicyRef:    policyRef,
			ReceiptCount: count,
			PolicyExists: policyExists,
			LastSeen:     policyLastSeen[policyRef].Format(time.RFC3339),
		})
	}

	// Step 4: Determine overall continuity status
	result.ContinuityOK = result.OrphanRefs == 0 && result.InvalidRefs == 0

	return result, nil
}

// validatePolicyExists checks if a policy reference exists in the policy system
func validatePolicyExists(ctx context.Context, pool *pgxpool.Pool, policyRef string) (bool, error) {
	// First try to find it in the canon table (loaded policies)
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM canon
		WHERE type LIKE 'policy%'
		AND (content::text ILIKE $1 OR type ILIKE $2)
	`, "%"+policyRef+"%", "%"+policyRef+"%").Scan(&count)

	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	// Also check if it's a standard policy name (gates.rego, risk.rego, etc.)
	standardPolicies := []string{"gates.rego", "risk.rego", "freeze.rego"}
	for _, standard := range standardPolicies {
		if policyRef == standard || policyRef == "policy."+standard {
			return true, nil
		}
	}

	return false, nil
}

// GetDomainReceiptStats returns receipt statistics for a specific domain
func GetDomainReceiptStats(ctx context.Context, pool *pgxpool.Pool, domainID string) (DomainReceiptStats, error) {
	stats := DomainReceiptStats{
		DomainRef: domainID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	domainPattern := "%" + domainID + "%"

	// Count total receipts for this domain
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM receipts
		WHERE event_id ILIKE $1 OR policy_ref ILIKE $1
	`, domainPattern).Scan(&stats.Total)
	if err != nil {
		return stats, fmt.Errorf("failed to count total receipts: %w", err)
	}

	// Count receipts with valid policy refs
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM receipts
		WHERE (event_id ILIKE $1 OR policy_ref ILIKE $1)
		AND policy_ref IS NOT NULL
		AND policy_ref != ''
	`, domainPattern).Scan(&stats.WithPolicyRef)
	if err != nil {
		return stats, fmt.Errorf("failed to count receipts with policy refs: %w", err)
	}

	// Count orphan receipts
	stats.Orphans = stats.Total - stats.WithPolicyRef

	// Get receipt type breakdown
	rows, err := pool.Query(ctx, `
		SELECT receipt_type, COUNT(*)
		FROM receipts
		WHERE event_id ILIKE $1 OR policy_ref ILIKE $1
		GROUP BY receipt_type
	`, domainPattern)
	if err != nil {
		return stats, fmt.Errorf("failed to get receipt type breakdown: %w", err)
	}
	defer rows.Close()

	stats.ByType = make(map[string]int)
	for rows.Next() {
		var receiptType string
		var count int
		if err := rows.Scan(&receiptType, &count); err != nil {
			continue
		}
		stats.ByType[receiptType] = count
	}

	return stats, nil
}

// DomainReceiptStats contains statistics for receipts associated with a domain
type DomainReceiptStats struct {
	DomainRef     string         `json:"domain_ref"`
	Total         int            `json:"total"`
	WithPolicyRef int            `json:"with_policy_ref"`
	Orphans       int            `json:"orphans"`
	ByType        map[string]int `json:"by_type"`
	Timestamp     string         `json:"timestamp"`
}

// RemediationResult represents the result of orphan remediation
type RemediationResult struct {
	Timestamp        string            `json:"timestamp"`
	TotalOrphans     int               `json:"total_orphans"`
	Remediated       int               `json:"remediated"`
	Failed           int               `json:"failed"`
	Strategies       []string          `json:"strategies_used"`
	Details          []RemediationItem `json:"details"`
	ContinuityBefore float64           `json:"continuity_before"`
	ContinuityAfter  float64           `json:"continuity_after"`
	ProofRefs        []string          `json:"proof_refs"`      // Phase 10F: Fix receipt IDs for lineage proofs
	LineageEnabled   bool              `json:"lineage_enabled"` // Phase 10F: Whether lineage proofs were created
}

// RemediationItem represents a single remediation action
type RemediationItem struct {
	ReceiptID    string `json:"receipt_id"`
	EventID      string `json:"event_id"`
	OldPolicyRef string `json:"old_policy_ref"`
	NewPolicyRef string `json:"new_policy_ref"`
	Strategy     string `json:"strategy"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	ProofRef     string `json:"proof_ref,omitempty"` // Phase 10F: Fix receipt ID for lineage proof
}

// ContinuityThresholds defines acceptable continuity levels
type ContinuityThresholds struct {
	Critical float64 `json:"critical"` // Below this is critical
	Warning  float64 `json:"warning"`  // Below this is warning
	Healthy  float64 `json:"healthy"`  // Above this is healthy
}

// DefaultContinuityThresholds returns standard continuity thresholds
func DefaultContinuityThresholds() ContinuityThresholds {
	return ContinuityThresholds{
		Critical: 50.0, // Below 50% is critical
		Warning:  75.0, // Below 75% is warning
		Healthy:  90.0, // Above 90% is healthy
	}
}

// RemediateOrphans attempts to fix orphaned receipts by inferring policy references
func RemediateOrphans(ctx context.Context, pool *pgxpool.Pool, domainID string, dryRun bool) (RemediationResult, error) {
	result := RemediationResult{
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Strategies:     []string{},
		Details:        []RemediationItem{},
		ProofRefs:      []string{}, // Phase 10F: Track fix receipt IDs
		LineageEnabled: !dryRun,    // Phase 10F: Enable lineage proofs in live mode
	}

	// Get current continuity status
	continuity, err := VerifyPolicyContinuity(ctx, pool, domainID)
	if err != nil {
		return result, fmt.Errorf("failed to get initial continuity: %w", err)
	}

	result.TotalOrphans = continuity.OrphanRefs
	result.ContinuityBefore = calculateContinuityRate(continuity.ValidRefs, continuity.TotalReceipts)

	if result.TotalOrphans == 0 {
		result.ContinuityAfter = result.ContinuityBefore
		return result, nil
	}

	// Strategy 1: Event ID pattern matching
	strategy1Count := 0
	for _, orphan := range continuity.OrphanReceipts {
		if orphan.PolicyRef != "" {
			continue // Skip if it has a policy ref but it's invalid
		}

		inferredPolicy := inferPolicyFromEventID(orphan.EventID, domainID)
		if inferredPolicy != "" {
			item := RemediationItem{
				ReceiptID:    orphan.ReceiptID,
				EventID:      orphan.EventID,
				OldPolicyRef: orphan.PolicyRef,
				NewPolicyRef: inferredPolicy,
				Strategy:     "event-id-pattern",
			}

			if !dryRun {
				err := updateReceiptPolicyRef(ctx, pool, orphan.ReceiptID, inferredPolicy)
				if err != nil {
					item.Success = false
					item.Error = err.Error()
					result.Failed++
				} else {
					item.Success = true
					result.Remediated++
					strategy1Count++

					// Phase 10F: Create fix receipt for lineage proof
					if fixID, fixErr := createFixReceipt(ctx, pool, orphan.ReceiptID, domainID, inferredPolicy, "event-id-pattern", "system"); fixErr == nil {
						item.ProofRef = fixID
						result.ProofRefs = append(result.ProofRefs, fixID)
					}
				}
			} else {
				item.Success = true // Assume success in dry run
				result.Remediated++
				strategy1Count++
			}

			result.Details = append(result.Details, item)
		}
	}

	if strategy1Count > 0 {
		result.Strategies = append(result.Strategies, fmt.Sprintf("event-id-pattern (%d receipts)", strategy1Count))
	}

	// Strategy 2: Domain default policy
	strategy2Count := 0
	defaultPolicy := inferDefaultPolicy(domainID)
	if defaultPolicy != "" {
		for _, orphan := range continuity.OrphanReceipts {
			if orphan.PolicyRef != "" {
				continue // Skip if it has a policy ref
			}

			// Only apply default if no other strategy worked
			alreadyHandled := false
			for _, detail := range result.Details {
				if detail.ReceiptID == orphan.ReceiptID {
					alreadyHandled = true
					break
				}
			}

			if !alreadyHandled {
				item := RemediationItem{
					ReceiptID:    orphan.ReceiptID,
					EventID:      orphan.EventID,
					OldPolicyRef: orphan.PolicyRef,
					NewPolicyRef: defaultPolicy,
					Strategy:     "domain-default",
				}

				if !dryRun {
					err := updateReceiptPolicyRef(ctx, pool, orphan.ReceiptID, defaultPolicy)
					if err != nil {
						item.Success = false
						item.Error = err.Error()
						result.Failed++
					} else {
						item.Success = true
						result.Remediated++
						strategy2Count++

						// Phase 10F: Create fix receipt for lineage proof
						if fixID, fixErr := createFixReceipt(ctx, pool, orphan.ReceiptID, domainID, defaultPolicy, "domain-default", "system"); fixErr == nil {
							item.ProofRef = fixID
							result.ProofRefs = append(result.ProofRefs, fixID)
						}
					}
				} else {
					item.Success = true
					result.Remediated++
					strategy2Count++
				}

				result.Details = append(result.Details, item)
			}
		}
	}

	if strategy2Count > 0 {
		result.Strategies = append(result.Strategies, fmt.Sprintf("domain-default (%d receipts)", strategy2Count))
	}

	// Calculate final continuity rate
	if !dryRun {
		updatedContinuity, err := VerifyPolicyContinuity(ctx, pool, domainID)
		if err != nil {
			result.ContinuityAfter = result.ContinuityBefore
		} else {
			result.ContinuityAfter = calculateContinuityRate(updatedContinuity.ValidRefs, updatedContinuity.TotalReceipts)
		}
	} else {
		// Estimate improvement for dry run
		totalAfter := continuity.TotalReceipts
		validAfter := continuity.ValidRefs + result.Remediated
		result.ContinuityAfter = calculateContinuityRate(validAfter, totalAfter)
	}

	return result, nil
}

// inferPolicyFromEventID attempts to infer policy reference from event ID patterns
func inferPolicyFromEventID(eventID, domainID string) string {
	// Pattern 1: event.domain.action -> policy.domain.action
	if len(eventID) > 6 && eventID[:6] == "event." {
		return "policy." + eventID[6:]
	}

	// Pattern 2: domain-specific events
	if eventID == "ci.call.v1" || eventID == "ci.import.v1" {
		return "policy.ci.validation.v1"
	}

	// Pattern 3: Basic domain mapping
	if domainID != "" {
		return fmt.Sprintf("policy.%s.default", domainID)
	}

	return ""
}

// inferDefaultPolicy returns a default policy for a domain
func inferDefaultPolicy(domainID string) string {
	if domainID == "" {
		return "policy.system.default"
	}
	return fmt.Sprintf("policy.%s.default", domainID)
}

// updateReceiptPolicyRef updates the policy reference for a receipt
func updateReceiptPolicyRef(ctx context.Context, pool *pgxpool.Pool, receiptID, policyRef string) error {
	query := `
		UPDATE receipts
		SET policy_ref = $1
		WHERE id = $2
	`

	_, err := pool.Exec(ctx, query, policyRef, receiptID)
	return err
}

// calculateContinuityRate calculates the continuity rate as a percentage
func calculateContinuityRate(valid, total int) float64 {
	if total == 0 {
		return 100.0
	}
	return float64(valid) / float64(total) * 100.0
}

// GetContinuityRiskLevel returns the risk level based on continuity rate
func GetContinuityRiskLevel(rate float64) string {
	thresholds := DefaultContinuityThresholds()

	if rate < thresholds.Critical {
		return "critical"
	} else if rate < thresholds.Warning {
		return "warning"
	} else if rate >= thresholds.Healthy {
		return "healthy"
	}
	return "acceptable"
}

// createFixReceipt creates a fix receipt for Phase 10F lineage proofs
func createFixReceipt(ctx context.Context, pool *pgxpool.Pool, originalReceiptID, domainID, policyRef, strategy, authorizedBy string) (string, error) {
	// Generate fix receipt ID with rcpt-fix- prefix
	fixID := "rcpt-fix-" + fmt.Sprintf("%d", time.Now().UnixNano())

	// Determine fix method based on strategy
	var fixMethod string
	switch strategy {
	case "event-id-pattern":
		fixMethod = FixMethodPatternMatch
	case "domain-default":
		fixMethod = FixMethodDomainDefault
	default:
		fixMethod = FixMethodManual
	}

	// Insert fix receipt
	_, err := pool.Exec(ctx, `
		INSERT INTO fix_receipts (id, original_receipt, domain_ref, policy_ref, fix_method, authorized_by, timestamp, verification)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, fixID, originalReceiptID, domainID, policyRef, fixMethod, authorizedBy, time.Now(), VerificationPending)

	if err != nil {
		return "", fmt.Errorf("failed to create fix receipt: %w", err)
	}

	return fixID, nil
}
