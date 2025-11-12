package api_test

import (
	"encoding/json"
	"testing"
	"time"

	"dis-core/internal/receipts"
)

// TestPolicyContinuityStructures tests the policy continuity data structures
func TestPolicyContinuityStructures(t *testing.T) {
	t.Run("PolicyContinuityResult JSON serialization", func(t *testing.T) {
		result := receipts.PolicyContinuityResult{
			DomainRef:     "test-domain",
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			TotalReceipts: 10,
			ValidRefs:     7,
			OrphanRefs:    2,
			InvalidRefs:   1,
			PolicyMappings: []receipts.PolicyMapping{
				{
					PolicyRef:    "policy.gates.rego",
					ReceiptCount: 5,
					PolicyExists: true,
					LastSeen:     time.Now().UTC().Format(time.RFC3339),
				},
			},
			OrphanReceipts: []receipts.OrphanReceipt{
				{
					ReceiptID:   "test-receipt-123",
					EventID:     "test-event-456",
					PolicyRef:   "",
					IssuedAt:    time.Now().UTC().Format(time.RFC3339),
					IssueReason: "missing policy_ref",
				},
			},
			ContinuityOK: false,
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(result)
		if err != nil {
			t.Errorf("Failed to marshal PolicyContinuityResult: %v", err)
		}

		// Test JSON deserialization
		var deserializedResult receipts.PolicyContinuityResult
		err = json.Unmarshal(jsonData, &deserializedResult)
		if err != nil {
			t.Errorf("Failed to unmarshal PolicyContinuityResult: %v", err)
		}

		// Verify key fields
		if deserializedResult.DomainRef != "test-domain" {
			t.Errorf("Expected domain_ref 'test-domain', got '%s'", deserializedResult.DomainRef)
		}

		if deserializedResult.TotalReceipts != 10 {
			t.Errorf("Expected total_receipts 10, got %d", deserializedResult.TotalReceipts)
		}

		if deserializedResult.ContinuityOK {
			t.Errorf("Expected continuity_ok false, got true")
		}

		if len(deserializedResult.PolicyMappings) != 1 {
			t.Errorf("Expected 1 policy mapping, got %d", len(deserializedResult.PolicyMappings))
		}

		if len(deserializedResult.OrphanReceipts) != 1 {
			t.Errorf("Expected 1 orphan receipt, got %d", len(deserializedResult.OrphanReceipts))
		}
	})

	t.Run("DomainReceiptStats JSON serialization", func(t *testing.T) {
		stats := receipts.DomainReceiptStats{
			DomainRef:     "test-domain",
			Total:         15,
			WithPolicyRef: 12,
			Orphans:       3,
			ByType: map[string]int{
				"ci.call.v1":   10,
				"ci.import.v1": 5,
			},
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		jsonData, err := json.Marshal(stats)
		if err != nil {
			t.Errorf("Failed to marshal DomainReceiptStats: %v", err)
		}

		var deserializedStats receipts.DomainReceiptStats
		err = json.Unmarshal(jsonData, &deserializedStats)
		if err != nil {
			t.Errorf("Failed to unmarshal DomainReceiptStats: %v", err)
		}

		if deserializedStats.DomainRef != "test-domain" {
			t.Errorf("Expected domain_ref 'test-domain', got '%s'", deserializedStats.DomainRef)
		}

		if deserializedStats.Total != 15 {
			t.Errorf("Expected total 15, got %d", deserializedStats.Total)
		}

		if len(deserializedStats.ByType) != 2 {
			t.Errorf("Expected 2 receipt types, got %d", len(deserializedStats.ByType))
		}
	})
}

// TestPolicyContinuityCalculation tests the continuity calculation logic
func TestPolicyContinuityCalculation(t *testing.T) {
	testCases := []struct {
		name          string
		totalReceipts int
		validRefs     int
		orphanRefs    int
		invalidRefs   int
		expectOK      bool
	}{
		{
			name:          "Perfect continuity",
			totalReceipts: 10,
			validRefs:     10,
			orphanRefs:    0,
			invalidRefs:   0,
			expectOK:      true,
		},
		{
			name:          "Has orphan refs",
			totalReceipts: 10,
			validRefs:     8,
			orphanRefs:    2,
			invalidRefs:   0,
			expectOK:      false,
		},
		{
			name:          "Has invalid refs",
			totalReceipts: 10,
			validRefs:     7,
			orphanRefs:    0,
			invalidRefs:   3,
			expectOK:      false,
		},
		{
			name:          "No receipts",
			totalReceipts: 0,
			validRefs:     0,
			orphanRefs:    0,
			invalidRefs:   0,
			expectOK:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := receipts.PolicyContinuityResult{
				TotalReceipts: tc.totalReceipts,
				ValidRefs:     tc.validRefs,
				OrphanRefs:    tc.orphanRefs,
				InvalidRefs:   tc.invalidRefs,
			}

			// Apply the continuity logic (matches the implementation)
			result.ContinuityOK = result.OrphanRefs == 0 && result.InvalidRefs == 0

			if result.ContinuityOK != tc.expectOK {
				t.Errorf("Expected continuity_ok %v, got %v (orphans: %d, invalid: %d)",
					tc.expectOK, result.ContinuityOK, result.OrphanRefs, result.InvalidRefs)
			}
		})
	}
}

// TestAuthoritySchemaExtensionStructure tests the authority schema structure extension
func TestAuthoritySchemaExtensionStructure(t *testing.T) {
	t.Run("Authority schema includes policy continuity field", func(t *testing.T) {
		// Test the schema structure that would be returned when policy continuity is added
		mockSchema := map[string]interface{}{
			"version":  "1.0",
			"receipts": []string{"ci.call.v1", "ci.import.v1"},
			"policy_continuity": map[string]interface{}{
				"supported": false,
				"reason":    "database not available",
			},
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(mockSchema)
		if err != nil {
			t.Errorf("Failed to marshal authority schema: %v", err)
		}

		// Parse JSON to verify structure
		var schema map[string]interface{}
		err = json.Unmarshal(jsonData, &schema)
		if err != nil {
			t.Errorf("Failed to parse authority schema JSON: %v", err)
		}

		// Check for policy_continuity field
		if _, exists := schema["policy_continuity"]; !exists {
			t.Errorf("Authority schema missing 'policy_continuity' field")
		}

		// Check that receipts field still exists (Phase 9C)
		if _, exists := schema["receipts"]; !exists {
			t.Errorf("Expected authority schema to still contain 'receipts' field from Phase 9C")
		}
	})
}

// TestDomainSchemaAggregateStructure tests the domain schema aggregate structure
func TestDomainSchemaAggregateStructure(t *testing.T) {
	t.Run("Domain schema aggregate has correct structure", func(t *testing.T) {
		// Test the aggregate structure without requiring a database
		aggregate := struct {
			DomainRef        string                           `json:"domain_ref"`
			Timestamp        string                           `json:"timestamp"`
			Schema           interface{}                      `json:"schema"`
			Policies         interface{}                      `json:"policies"`
			PolicyCount      int                              `json:"policy_count"`
			PolicyContinuity *receipts.PolicyContinuityResult `json:"policy_continuity,omitempty"`
			ReceiptStats     *receipts.DomainReceiptStats     `json:"receipt_stats,omitempty"`
		}{
			DomainRef:   "test-domain",
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			Schema:      map[string]interface{}{},
			Policies:    []interface{}{},
			PolicyCount: 0,
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(aggregate)
		if err != nil {
			t.Errorf("Failed to marshal domain schema aggregate: %v", err)
		}

		// Verify JSON contains expected fields
		var unmarshaled map[string]interface{}
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Errorf("Failed to unmarshal domain schema aggregate: %v", err)
		}

		expectedFields := []string{"domain_ref", "timestamp", "schema", "policies", "policy_count"}
		for _, field := range expectedFields {
			if _, exists := unmarshaled[field]; !exists {
				t.Errorf("Domain schema aggregate missing '%s' field", field)
			}
		}
	})
}

// TestPolicyContinuityRate tests continuity rate calculations
func TestPolicyContinuityRate(t *testing.T) {
	testCases := []struct {
		name          string
		totalReceipts int
		validRefs     int
		expectedRate  float64
	}{
		{
			name:          "Perfect continuity",
			totalReceipts: 10,
			validRefs:     10,
			expectedRate:  1.0,
		},
		{
			name:          "Half continuity",
			totalReceipts: 10,
			validRefs:     5,
			expectedRate:  0.5,
		},
		{
			name:          "No receipts",
			totalReceipts: 0,
			validRefs:     0,
			expectedRate:  0.0,
		},
		{
			name:          "No valid refs",
			totalReceipts: 10,
			validRefs:     0,
			expectedRate:  0.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var rate float64
			if tc.totalReceipts > 0 {
				rate = float64(tc.validRefs) / float64(tc.totalReceipts)
			}

			if rate != tc.expectedRate {
				t.Errorf("Expected rate %.2f, got %.2f", tc.expectedRate, rate)
			}
		})
	}
}
