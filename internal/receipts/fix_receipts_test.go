package receipts

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFixReceiptTypes tests the core Phase 10F data structures
func TestFixReceiptTypes(t *testing.T) {
	t.Run("FixReceipt creation", func(t *testing.T) {
		receipt := FixReceipt{
			ID:              "rcpt-fix-test-001",
			OriginalReceipt: "550e8400-e29b-41d4-a716-446655440000",
			DomainRef:       "domain.test.v1",
			ActionRef:       "action.test.v1",
			PolicyRef:       "policy.test.v1",
			FixMethod:       FixMethodPatternMatch,
			AuthorizedBy:    "authority.test",
			Timestamp:       time.Now().UTC(),
			Verification:    VerificationPending,
		}

		assert.True(t, strings.HasPrefix(receipt.ID, "rcpt-fix-"))
		assert.Equal(t, FixMethodPatternMatch, receipt.FixMethod)
		assert.Equal(t, VerificationPending, receipt.Verification)
		assert.NotEmpty(t, receipt.DomainRef)
	})

	t.Run("LineageProof structure", func(t *testing.T) {
		now := time.Now().UTC()
		proof := LineageProof{
			OriginalReceiptID: "550e8400-e29b-41d4-a716-446655440000",
			FixReceiptID:      "rcpt-fix-test-001",
			ProofChain:        []string{"proof-001", "proof-002"},
			Verified:          true,
			CreatedAt:         now,
			VerifiedAt:        &now,
		}

		assert.Equal(t, "rcpt-fix-test-001", proof.FixReceiptID)
		assert.True(t, proof.Verified)
		assert.Len(t, proof.ProofChain, 2)
		assert.NotNil(t, proof.VerifiedAt)
	})

	t.Run("LineageSummary calculations", func(t *testing.T) {
		summary := LineageSummary{
			TotalFixReceipts:     10,
			PendingVerifications: 3,
			VerifiedChains:       7,
			VerifiedChainRate:    70.0,
			LastFixTimestamp:     time.Now().UTC().Format(time.RFC3339),
			RecentFixes:          []FixReceipt{},
		}

		assert.Equal(t, 10, summary.TotalFixReceipts)
		assert.Equal(t, 3, summary.PendingVerifications)
		assert.Equal(t, 7, summary.VerifiedChains)
		assert.Equal(t, 70.0, summary.VerifiedChainRate)
		assert.NotEmpty(t, summary.LastFixTimestamp)
	})
}

// TestFixReceiptJSONSerialization tests JSON marshaling/unmarshaling
func TestFixReceiptJSONSerialization(t *testing.T) {
	originalReceipt := FixReceipt{
		ID:              "rcpt-fix-json-test",
		OriginalReceipt: "550e8400-e29b-41d4-a716-446655440000",
		DomainRef:       "domain.serialization.v1",
		ActionRef:       "action.serialization.v1",
		PolicyRef:       "policy.serialization.v1",
		FixMethod:       FixMethodManual,
		AuthorizedBy:    "authority.serialization",
		Timestamp:       time.Now().UTC().Truncate(time.Second),
		Verification:    VerificationHash,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(originalReceipt)
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), "rcpt-fix-json-test")
	assert.Contains(t, string(jsonData), FixMethodManual)

	// Unmarshal from JSON
	var deserializedReceipt FixReceipt
	err = json.Unmarshal(jsonData, &deserializedReceipt)
	require.NoError(t, err)

	assert.Equal(t, originalReceipt.ID, deserializedReceipt.ID)
	assert.Equal(t, originalReceipt.FixMethod, deserializedReceipt.FixMethod)
	assert.Equal(t, originalReceipt.DomainRef, deserializedReceipt.DomainRef)
	assert.Equal(t, originalReceipt.Verification, deserializedReceipt.Verification)
}

// Integration tests requiring database connection would go in a separate file
// with build tags to allow running unit tests without database setup

/*
// Example database integration test structure (would be in separate file)

func TestFixReceiptDatabaseOperations(t *testing.T) {
	// This would test actual database operations
	// Requires test database setup and connection
	pool := setupTestDatabase(t)
	defer pool.Close()

	t.Run("Create fix receipt", func(t *testing.T) {
		// Test createFixReceipt function
	})

	t.Run("Query lineage summary", func(t *testing.T) {
		// Test getLineageSummary function
	})

	t.Run("Update verification status", func(t *testing.T) {
		// Test verification status updates
	})
}
*/
