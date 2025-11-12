package receipts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ProofSynchronizer handles cross-domain proof synchronization and verification
type ProofSynchronizer struct {
	dbPool         *pgxpool.Pool
	localDomain    string
	trustedDomains map[string]string // domain -> trust_level mapping
}

// NewProofSynchronizer creates a new proof synchronizer instance
func NewProofSynchronizer(dbPool *pgxpool.Pool, localDomain string) *ProofSynchronizer {
	return &ProofSynchronizer{
		dbPool:         dbPool,
		localDomain:    localDomain,
		trustedDomains: make(map[string]string),
	}
}

// SyncProofs synchronizes lineage proofs between domains
func (ps *ProofSynchronizer) SyncProofs(ctx context.Context, sourceDomain, targetDomain string) (*ProofSyncResponse, error) {
	response := &ProofSyncResponse{
		Status:        "success",
		SyncedProofs:  0,
		FailedProofs:  0,
		Discrepancies: []string{},
		Timestamp:     time.Now().UTC(),
		Details:       make(map[string]string),
	}

	// Step 1: Validate domain trust relationship
	trustLevel, err := ps.getTrustLevel(ctx, sourceDomain, targetDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to get trust level: %w", err)
	}

	if trustLevel == TrustLevelNone {
		response.Status = "failed"
		response.Details["error"] = "No trust relationship between domains"
		return response, nil
	}

	// Step 2: Get proofs that need synchronization
	proofs, err := ps.getProofsForSync(ctx, sourceDomain, targetDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to get proofs for sync: %w", err)
	}

	// Step 3: Process each proof
	for _, proof := range proofs {
		if err := ps.syncSingleProof(ctx, proof, targetDomain, trustLevel); err != nil {
			response.FailedProofs++
			response.Discrepancies = append(response.Discrepancies,
				fmt.Sprintf("Proof %s: %v", proof.FixReceiptID, err))
		} else {
			response.SyncedProofs++
		}
	}

	// Step 4: Update sync metadata
	response.Details["trust_level"] = trustLevel
	response.Details["total_processed"] = fmt.Sprintf("%d", len(proofs))

	return response, nil
}

// VerifyForeignProof validates a proof from another domain
func (ps *ProofSynchronizer) VerifyForeignProof(ctx context.Context, proofID string) (bool, error) {
	// Step 1: Retrieve the federation proof record
	var fedProof FederationProof
	query := `
		SELECT id, source_domain, target_domain, proof_ref, federation_hash,
		       status, trust_level, timestamp, verified_at
		FROM federation_proofs
		WHERE id = $1 OR proof_ref = $1
	`

	row := ps.dbPool.QueryRow(ctx, query, proofID)
	err := row.Scan(&fedProof.ID, &fedProof.SourceDomain, &fedProof.TargetDomain,
		&fedProof.ProofRef, &fedProof.FederationHash, &fedProof.Status,
		&fedProof.TrustLevel, &fedProof.Timestamp, &fedProof.VerifiedAt)

	if err != nil {
		return false, fmt.Errorf("federation proof not found: %w", err)
	}

	// Step 2: Recalculate federation hash for verification
	originalProof, err := ps.getOriginalProof(ctx, fedProof.ProofRef)
	if err != nil {
		return false, fmt.Errorf("failed to get original proof: %w", err)
	}

	expectedHash := ps.CalculateFederationHash(*originalProof)

	// Step 3: Compare hashes
	if expectedHash != fedProof.FederationHash {
		// Mark as rejected due to hash mismatch
		if err := ps.updateProofStatus(ctx, fedProof.ID, CrossDomainStatusRejected); err != nil {
			return false, fmt.Errorf("failed to update proof status: %w", err)
		}
		return false, fmt.Errorf("hash mismatch: expected %s, got %s",
			expectedHash, fedProof.FederationHash)
	}

	// Step 4: Mark as verified if hash matches
	now := time.Now().UTC()
	if err := ps.markProofVerified(ctx, fedProof.ID, now); err != nil {
		return false, fmt.Errorf("failed to mark proof as verified: %w", err)
	}

	return true, nil
}

// CalculateFederationHash generates a hash for cross-domain proof verification
func (ps *ProofSynchronizer) CalculateFederationHash(proof LineageProof) string {
	// Create a canonical representation of the proof for hashing
	hashData := struct {
		OriginalReceiptID string   `json:"original_receipt_id"`
		FixReceiptID      string   `json:"fix_receipt_id"`
		ProofChain        []string `json:"proof_chain"`
		SourceDomain      string   `json:"source_domain"`
		CreatedAt         string   `json:"created_at"`
	}{
		OriginalReceiptID: proof.OriginalReceiptID,
		FixReceiptID:      proof.FixReceiptID,
		ProofChain:        proof.ProofChain,
		SourceDomain:      proof.SourceDomain,
		CreatedAt:         proof.CreatedAt.UTC().Format(time.RFC3339),
	}

	// Serialize to JSON for consistent hashing
	jsonData, err := json.Marshal(hashData)
	if err != nil {
		// Fallback to simple concatenation if JSON marshaling fails
		combined := proof.OriginalReceiptID + proof.FixReceiptID + proof.SourceDomain
		hash := sha256.Sum256([]byte(combined))
		return hex.EncodeToString(hash[:])
	}

	// Calculate SHA256 hash
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// GetFederationSummary provides dashboard metrics for cross-domain verification
func (ps *ProofSynchronizer) GetFederationSummary(ctx context.Context) (*FederationSummary, error) {
	summary := &FederationSummary{
		TrustedDomains: []string{},
	}

	// Get total federation proof counts
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'verified' THEN 1 END) as verified,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
			COUNT(CASE WHEN status = 'rejected' OR status = 'conflict' THEN 1 END) as discrepancies,
			MAX(timestamp) as last_sync
		FROM federation_proofs
	`

	row := ps.dbPool.QueryRow(ctx, query)
	var lastSync *time.Time
	err := row.Scan(&summary.TotalFederationProofs, &summary.VerifiedProofs,
		&summary.PendingProofs, &summary.DiscrepancyCount, &lastSync)

	if err != nil {
		return nil, fmt.Errorf("failed to get federation summary: %w", err)
	}

	// Calculate verification rate
	if summary.TotalFederationProofs > 0 {
		summary.VerificationRate = float64(summary.VerifiedProofs) / float64(summary.TotalFederationProofs) * 100
	}

	// Set last sync timestamp
	if lastSync != nil {
		summary.LastSyncTimestamp = lastSync.UTC().Format(time.RFC3339)
	}

	// Get trusted domains
	trustedQuery := `
		SELECT DISTINCT domain_a, domain_b
		FROM federation_trust
		WHERE trust_level IN ('high', 'medium')
		ORDER BY domain_a, domain_b
	`

	rows, err := ps.dbPool.Query(ctx, trustedQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get trusted domains: %w", err)
	}
	defer rows.Close()

	domainSet := make(map[string]bool)
	for rows.Next() {
		var domainA, domainB string
		if err := rows.Scan(&domainA, &domainB); err != nil {
			continue
		}
		domainSet[domainA] = true
		domainSet[domainB] = true
	}

	for domain := range domainSet {
		if domain != ps.localDomain {
			summary.TrustedDomains = append(summary.TrustedDomains, domain)
		}
	}

	return summary, nil
}

// Helper functions

func (ps *ProofSynchronizer) getTrustLevel(ctx context.Context, domainA, domainB string) (string, error) {
	query := `
		SELECT trust_level
		FROM federation_trust
		WHERE (domain_a = $1 AND domain_b = $2) OR (domain_a = $2 AND domain_b = $1)
		ORDER BY trust_level DESC
		LIMIT 1
	`

	var trustLevel string
	err := ps.dbPool.QueryRow(ctx, query, domainA, domainB).Scan(&trustLevel)
	if err != nil {
		return TrustLevelNone, nil // No trust relationship found
	}

	return trustLevel, nil
}

func (ps *ProofSynchronizer) getProofsForSync(ctx context.Context, sourceDomain, targetDomain string) ([]LineageProof, error) {
	// This would typically query the lineage_proofs table or fix_receipts
	// For now, return an empty slice as this would depend on the specific sync strategy
	return []LineageProof{}, nil
}

func (ps *ProofSynchronizer) syncSingleProof(ctx context.Context, proof LineageProof, targetDomain, trustLevel string) error {
	// Calculate federation hash
	federationHash := ps.CalculateFederationHash(proof)

	// Create federation proof record
	insertQuery := `
		INSERT INTO federation_proofs (
			id, source_domain, target_domain, proof_ref, federation_hash,
			status, trust_level, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			federation_hash = $5,
			status = $6,
			timestamp = $8
	`

	federationProofID := fmt.Sprintf("fed-proof-%s-%s", proof.FixReceiptID, targetDomain)
	_, err := ps.dbPool.Exec(ctx, insertQuery,
		federationProofID, proof.SourceDomain, targetDomain, proof.FixReceiptID,
		federationHash, CrossDomainStatusPending, trustLevel, time.Now().UTC())

	return err
}

func (ps *ProofSynchronizer) getOriginalProof(ctx context.Context, proofRef string) (*LineageProof, error) {
	// This would query the actual lineage proof from the database
	// For now, return a placeholder
	return &LineageProof{
		FixReceiptID: proofRef,
		// Other fields would be populated from database
	}, nil
}

func (ps *ProofSynchronizer) updateProofStatus(ctx context.Context, proofID, status string) error {
	query := `UPDATE federation_proofs SET status = $1 WHERE id = $2`
	_, err := ps.dbPool.Exec(ctx, query, status, proofID)
	return err
}

func (ps *ProofSynchronizer) markProofVerified(ctx context.Context, proofID string, verifiedAt time.Time) error {
	query := `
		UPDATE federation_proofs
		SET status = $1, verified_at = $2
		WHERE id = $3
	`
	_, err := ps.dbPool.Exec(ctx, query, CrossDomainStatusVerified, verifiedAt, proofID)
	return err
}
