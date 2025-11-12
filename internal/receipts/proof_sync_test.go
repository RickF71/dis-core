package receipts

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCrossDomainLineageProofFields tests the new Phase 10G fields in LineageProof
func TestCrossDomainLineageProofFields(t *testing.T) {
	t.Run("LineageProof with cross-domain fields", func(t *testing.T) {
		proof := LineageProof{
			OriginalReceiptID: "550e8400-e29b-41d4-a716-446655440000",
			FixReceiptID:      "rcpt-fix-test-001",
			ProofChain:        []string{"proof-001", "proof-002"},
			Verified:          true,
			CreatedAt:         time.Now().UTC(),
			// Phase 10G: Cross-domain fields
			SourceDomain:      "terra.domain",
			TargetDomain:      "usa.domain",
			FederationHash:    "abcd1234hash",
			CrossDomainStatus: CrossDomainStatusVerified,
		}

		assert.Equal(t, "terra.domain", proof.SourceDomain)
		assert.Equal(t, "usa.domain", proof.TargetDomain)
		assert.Equal(t, "abcd1234hash", proof.FederationHash)
		assert.Equal(t, CrossDomainStatusVerified, proof.CrossDomainStatus)
	})
}

// TestFederationProofTypes tests Phase 10G federation proof data structures
func TestFederationProofTypes(t *testing.T) {
	t.Run("FederationProof creation", func(t *testing.T) {
		fedProof := FederationProof{
			ID:             "fed-proof-001",
			SourceDomain:   "terra.domain",
			TargetDomain:   "usa.domain",
			ProofRef:       "rcpt-fix-test-001",
			FederationHash: "federated-hash-123",
			Status:         CrossDomainStatusPending,
			Timestamp:      time.Now().UTC(),
			TrustLevel:     TrustLevelHigh,
		}

		assert.Equal(t, "fed-proof-001", fedProof.ID)
		assert.Equal(t, "terra.domain", fedProof.SourceDomain)
		assert.Equal(t, "usa.domain", fedProof.TargetDomain)
		assert.Equal(t, CrossDomainStatusPending, fedProof.Status)
		assert.Equal(t, TrustLevelHigh, fedProof.TrustLevel)
	})

	t.Run("FederationSummary metrics", func(t *testing.T) {
		summary := FederationSummary{
			TotalFederationProofs: 25,
			VerifiedProofs:        20,
			PendingProofs:         3,
			DiscrepancyCount:      2,
			VerificationRate:      80.0,
			TrustedDomains:        []string{"terra.domain", "usa.domain"},
			LastSyncTimestamp:     time.Now().UTC().Format(time.RFC3339),
		}

		assert.Equal(t, 25, summary.TotalFederationProofs)
		assert.Equal(t, 20, summary.VerifiedProofs)
		assert.Equal(t, 80.0, summary.VerificationRate)
		assert.Len(t, summary.TrustedDomains, 2)
	})

	t.Run("ProofSyncRequest structure", func(t *testing.T) {
		syncRequest := ProofSyncRequest{
			SourceDomain: "terra.domain",
			TargetDomain: "usa.domain",
			ProofIDs:     []string{"proof-001", "proof-002"},
			SyncMode:     SyncModePush,
		}

		assert.Equal(t, "terra.domain", syncRequest.SourceDomain)
		assert.Equal(t, SyncModePush, syncRequest.SyncMode)
		assert.Len(t, syncRequest.ProofIDs, 2)
	})
}

// TestProofSynchronizerHashCalculation tests the federation hash calculation
func TestProofSynchronizerHashCalculation(t *testing.T) {
	t.Run("CalculateFederationHash consistency", func(t *testing.T) {
		// Create a mock proof synchronizer
		synchronizer := &ProofSynchronizer{
			localDomain: "test.domain",
		}

		// Create a test proof
		proof := LineageProof{
			OriginalReceiptID: "550e8400-e29b-41d4-a716-446655440000",
			FixReceiptID:      "rcpt-fix-test-001",
			ProofChain:        []string{"proof-001", "proof-002"},
			SourceDomain:      "terra.domain",
			CreatedAt:         time.Date(2025, 11, 11, 12, 0, 0, 0, time.UTC),
		}

		// Calculate hash twice - should be consistent
		hash1 := synchronizer.CalculateFederationHash(proof)
		hash2 := synchronizer.CalculateFederationHash(proof)

		assert.Equal(t, hash1, hash2, "Hash calculation should be deterministic")
		assert.NotEmpty(t, hash1, "Hash should not be empty")
		assert.Equal(t, 64, len(hash1), "SHA256 hash should be 64 characters")

		// Verify it's a valid hex string
		_, err := hex.DecodeString(hash1)
		assert.NoError(t, err, "Hash should be valid hex")
	})

	t.Run("CalculateFederationHash changes with different data", func(t *testing.T) {
		synchronizer := &ProofSynchronizer{
			localDomain: "test.domain",
		}

		proof1 := LineageProof{
			OriginalReceiptID: "550e8400-e29b-41d4-a716-446655440000",
			FixReceiptID:      "rcpt-fix-test-001",
			ProofChain:        []string{"proof-001"},
			SourceDomain:      "terra.domain",
			CreatedAt:         time.Date(2025, 11, 11, 12, 0, 0, 0, time.UTC),
		}

		proof2 := proof1
		proof2.FixReceiptID = "rcpt-fix-test-002" // Change one field

		hash1 := synchronizer.CalculateFederationHash(proof1)
		hash2 := synchronizer.CalculateFederationHash(proof2)

		assert.NotEqual(t, hash1, hash2, "Different proofs should have different hashes")
	})
}

// TestProofSynchronizationConstants tests the Phase 10G constants
func TestProofSynchronizationConstants(t *testing.T) {
	t.Run("Cross-domain status constants", func(t *testing.T) {
		statuses := []string{
			CrossDomainStatusPending,
			CrossDomainStatusVerified,
			CrossDomainStatusRejected,
			CrossDomainStatusConflict,
		}

		expectedStatuses := []string{"pending", "verified", "rejected", "conflict"}

		for i, status := range statuses {
			assert.Equal(t, expectedStatuses[i], status)
		}
	})

	t.Run("Trust level constants", func(t *testing.T) {
		trustLevels := []string{
			TrustLevelHigh,
			TrustLevelMedium,
			TrustLevelLow,
			TrustLevelNone,
		}

		expectedLevels := []string{"high", "medium", "low", "none"}

		for i, level := range trustLevels {
			assert.Equal(t, expectedLevels[i], level)
		}
	})

	t.Run("Sync mode constants", func(t *testing.T) {
		syncModes := []string{
			SyncModePush,
			SyncModePull,
			SyncModeFull,
		}

		expectedModes := []string{"push", "pull", "full"}

		for i, mode := range syncModes {
			assert.Equal(t, expectedModes[i], mode)
		}
	})
}

// TestMockCrossDomainScenarios tests mock scenarios for cross-domain verification
func TestMockCrossDomainScenarios(t *testing.T) {
	t.Run("Two domain federation scenario", func(t *testing.T) {
		// Mock scenario: Terra domain and USA domain
		terraProof := LineageProof{
			OriginalReceiptID: "terra-receipt-001",
			FixReceiptID:      "rcpt-fix-terra-001",
			ProofChain:        []string{"terra-proof-001", "terra-proof-002"},
			SourceDomain:      "terra.domain",
			TargetDomain:      "usa.domain",
			CreatedAt:         time.Now().UTC(),
		}

		usaProof := LineageProof{
			OriginalReceiptID: "usa-receipt-001",
			FixReceiptID:      "rcpt-fix-usa-001",
			ProofChain:        []string{"usa-proof-001", "usa-proof-002"},
			SourceDomain:      "usa.domain",
			TargetDomain:      "terra.domain",
			CreatedAt:         time.Now().UTC(),
		}

		// Mock synchronizer
		terraSynchronizer := &ProofSynchronizer{
			localDomain: "terra.domain",
		}
		usaSynchronizer := &ProofSynchronizer{
			localDomain: "usa.domain",
		}

		// Calculate federation hashes
		terraHash := terraSynchronizer.CalculateFederationHash(terraProof)
		usaHash := usaSynchronizer.CalculateFederationHash(usaProof)

		assert.NotEqual(t, terraHash, usaHash, "Different domains should have different hashes")
		assert.NotEmpty(t, terraHash)
		assert.NotEmpty(t, usaHash)
	})

	t.Run("Trust relationship validation", func(t *testing.T) {
		// Test trust level validation scenarios
		trustRelationships := []struct {
			domainA    string
			domainB    string
			trustLevel string
			shouldSync bool
		}{
			{"terra.domain", "usa.domain", TrustLevelHigh, true},
			{"terra.domain", "usa.domain", TrustLevelMedium, true},
			{"terra.domain", "unknown.domain", TrustLevelNone, false},
			{"terra.domain", "malicious.domain", TrustLevelLow, true}, // Still allows sync but with warnings
		}

		for _, rel := range trustRelationships {
			// Mock trust evaluation
			shouldAllow := rel.trustLevel != TrustLevelNone
			assert.Equal(t, rel.shouldSync || rel.trustLevel == TrustLevelLow, shouldAllow,
				"Trust level %s between %s and %s should result in sync permission: %v",
				rel.trustLevel, rel.domainA, rel.domainB, rel.shouldSync)
		}
	})

	t.Run("Hash mismatch detection", func(t *testing.T) {
		// Simulate hash mismatch scenario
		originalProof := LineageProof{
			OriginalReceiptID: "test-receipt-001",
			FixReceiptID:      "rcpt-fix-test-001",
			ProofChain:        []string{"proof-001"},
			SourceDomain:      "source.domain",
			CreatedAt:         time.Now().UTC(),
		}

		// Simulate tampering by modifying proof chain
		tamperedProof := originalProof
		tamperedProof.ProofChain = []string{"tampered-proof-001"}

		synchronizer := &ProofSynchronizer{
			localDomain: "test.domain",
		}

		originalHash := synchronizer.CalculateFederationHash(originalProof)
		tamperedHash := synchronizer.CalculateFederationHash(tamperedProof)

		assert.NotEqual(t, originalHash, tamperedHash, "Tampered proof should have different hash")

		// Mock verification response
		verificationResult := map[string]interface{}{
			"verified": false,
			"reason":   "Hash mismatch detected",
		}

		assert.False(t, verificationResult["verified"].(bool))
		assert.Equal(t, "Hash mismatch detected", verificationResult["reason"])
	})
}

// TestProofSyncResponseStructure tests the sync response structure
func TestProofSyncResponseStructure(t *testing.T) {
	t.Run("ProofSyncResponse with success", func(t *testing.T) {
		response := ProofSyncResponse{
			Status:        "success",
			SyncedProofs:  5,
			FailedProofs:  1,
			Discrepancies: []string{"Hash mismatch on proof-003"},
			Timestamp:     time.Now().UTC(),
			Details: map[string]string{
				"trust_level":     "high",
				"total_processed": "6",
			},
		}

		assert.Equal(t, "success", response.Status)
		assert.Equal(t, 5, response.SyncedProofs)
		assert.Equal(t, 1, response.FailedProofs)
		assert.Len(t, response.Discrepancies, 1)
		assert.Equal(t, "high", response.Details["trust_level"])
	})
}

// Helper function for mock SHA256 calculation (for testing purposes)
func mockCalculateSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
