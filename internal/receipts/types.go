package receipts

import "time"

// FixReceipt represents a rcpt.fix.v1 receipt type for continuity lineage proofs
type FixReceipt struct {
	ID              string    `json:"id"`
	OriginalReceipt string    `json:"original_receipt"`
	DomainRef       string    `json:"domain_ref"`
	ActionRef       string    `json:"action_ref"`
	PolicyRef       string    `json:"policy_ref"`
	FixMethod       string    `json:"fix_method"`
	AuthorizedBy    string    `json:"authorized_by"`
	Timestamp       time.Time `json:"timestamp"`
	Verification    string    `json:"verification"`
}

// FixMethod constants for different remediation strategies
const (
	FixMethodPatternMatch  = "pattern-match"
	FixMethodDomainDefault = "domain-default"
	FixMethodManual        = "manual"
)

// Verification status constants
const (
	VerificationPending   = "pending"
	VerificationSignature = "signature"
	VerificationHash      = "hash"
)

// LineageProof represents the complete lineage proof for a fix
type LineageProof struct {
	OriginalReceiptID string     `json:"original_receipt_id"`
	FixReceiptID      string     `json:"fix_receipt_id"`
	ProofChain        []string   `json:"proof_chain"`
	Verified          bool       `json:"verified"`
	CreatedAt         time.Time  `json:"created_at"`
	VerifiedAt        *time.Time `json:"verified_at,omitempty"`
	// Phase 10G: Cross-Domain Verification Fields
	SourceDomain      string `json:"source_domain,omitempty"`
	TargetDomain      string `json:"target_domain,omitempty"`
	FederationHash    string `json:"federation_hash,omitempty"`
	CrossDomainStatus string `json:"cross_domain_status,omitempty"`
}

// LineageSummary provides dashboard metrics for lineage proof system
type LineageSummary struct {
	TotalFixReceipts     int          `json:"total_fix_receipts"`
	PendingVerifications int          `json:"pending_verifications"`
	VerifiedChains       int          `json:"verified_chains"`
	VerifiedChainRate    float64      `json:"verified_chain_rate"`
	LastFixTimestamp     string       `json:"last_fix_timestamp,omitempty"`
	RecentFixes          []FixReceipt `json:"recent_fixes"`
}

// Phase 10G: Federation and Cross-Domain Proof Types

// FederationProof represents a cross-domain lineage proof
type FederationProof struct {
	ID             string     `json:"id"`
	SourceDomain   string     `json:"source_domain"`
	TargetDomain   string     `json:"target_domain"`
	ProofRef       string     `json:"proof_ref"`
	FederationHash string     `json:"federation_hash"`
	Status         string     `json:"status"`
	Timestamp      time.Time  `json:"timestamp"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty"`
	TrustLevel     string     `json:"trust_level"`
}

// FederationSummary provides dashboard metrics for cross-domain proof verification
type FederationSummary struct {
	TotalFederationProofs int      `json:"total_federation_proofs"`
	VerifiedProofs        int      `json:"verified_proofs"`
	PendingProofs         int      `json:"pending_proofs"`
	DiscrepancyCount      int      `json:"discrepancy_count"`
	VerificationRate      float64  `json:"verification_rate"`
	TrustedDomains        []string `json:"trusted_domains"`
	LastSyncTimestamp     string   `json:"last_sync_timestamp,omitempty"`
}

// ProofSyncRequest represents a request to synchronize proofs between domains
type ProofSyncRequest struct {
	SourceDomain string   `json:"source_domain"`
	TargetDomain string   `json:"target_domain"`
	ProofIDs     []string `json:"proof_ids"`
	SyncMode     string   `json:"sync_mode"` // "push" or "pull"
}

// ProofSyncResponse represents the response from a proof synchronization operation
type ProofSyncResponse struct {
	Status        string            `json:"status"`
	SyncedProofs  int               `json:"synced_proofs"`
	FailedProofs  int               `json:"failed_proofs"`
	Discrepancies []string          `json:"discrepancies"`
	Timestamp     time.Time         `json:"timestamp"`
	Details       map[string]string `json:"details"`
}

// Cross-domain proof status constants
const (
	CrossDomainStatusPending  = "pending"
	CrossDomainStatusVerified = "verified"
	CrossDomainStatusRejected = "rejected"
	CrossDomainStatusConflict = "conflict"
)

// Federation trust level constants
const (
	TrustLevelHigh   = "high"
	TrustLevelMedium = "medium"
	TrustLevelLow    = "low"
	TrustLevelNone   = "none"
)

// Sync mode constants
const (
	SyncModePush = "push"
	SyncModePull = "pull"
	SyncModeFull = "full"
)
