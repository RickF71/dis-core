package api_test

import (
	"testing"

	"dis-core/internal/receipts"
)

func TestPhase10EContinuityThresholds(t *testing.T) {
	thresholds := receipts.DefaultContinuityThresholds()

	if thresholds.Critical <= 0 || thresholds.Critical >= 100 {
		t.Errorf("Critical threshold should be between 0 and 100, got %.1f", thresholds.Critical)
	}

	if thresholds.Warning <= thresholds.Critical || thresholds.Warning >= 100 {
		t.Errorf("Warning threshold should be between critical and 100, got %.1f", thresholds.Warning)
	}

	if thresholds.Healthy <= thresholds.Warning || thresholds.Healthy > 100 {
		t.Errorf("Healthy threshold should be between warning and 100, got %.1f", thresholds.Healthy)
	}

	// Test risk level calculation
	tests := []struct {
		rate     float64
		expected string
	}{
		{30.0, "critical"},
		{60.0, "warning"},
		{85.0, "acceptable"},
		{95.0, "healthy"},
	}

	for _, tt := range tests {
		result := receipts.GetContinuityRiskLevel(tt.rate)
		if result != tt.expected {
			t.Errorf("For rate %.1f, expected risk level '%s', got '%s'",
				tt.rate, tt.expected, result)
		}
	}
}

func TestPhase10ERemediationStructures(t *testing.T) {
	// Test RemediationResult structure
	result := receipts.RemediationResult{
		Timestamp:        "2025-11-11T00:00:00Z",
		TotalOrphans:     10,
		Remediated:       8,
		Failed:           2,
		Strategies:       []string{"event-id-pattern (5 receipts)", "domain-default (3 receipts)"},
		ContinuityBefore: 70.0,
		ContinuityAfter:  95.0,
		Details: []receipts.RemediationItem{
			{
				ReceiptID:    "test-receipt-1",
				EventID:      "event.test.action",
				OldPolicyRef: "",
				NewPolicyRef: "policy.test.action",
				Strategy:     "event-id-pattern",
				Success:      true,
			},
		},
	}

	// Verify structure completeness
	if result.Timestamp == "" {
		t.Error("Expected timestamp to be set")
	}
	if len(result.Strategies) == 0 {
		t.Error("Expected strategies to be populated")
	}
	if len(result.Details) == 0 {
		t.Error("Expected details to be populated")
	}
	if result.ContinuityAfter <= result.ContinuityBefore {
		t.Error("Expected continuity to improve after remediation")
	}
}

func TestPhase10EThresholdsIntegrity(t *testing.T) {
	thresholds := receipts.DefaultContinuityThresholds()

	// Verify threshold ordering
	if thresholds.Critical >= thresholds.Warning {
		t.Error("Critical threshold must be less than warning threshold")
	}
	if thresholds.Warning >= thresholds.Healthy {
		t.Error("Warning threshold must be less than healthy threshold")
	}

	// Verify reasonable values
	expectedCritical := 50.0
	expectedWarning := 75.0
	expectedHealthy := 90.0

	if thresholds.Critical != expectedCritical {
		t.Errorf("Expected critical threshold %.1f, got %.1f", expectedCritical, thresholds.Critical)
	}
	if thresholds.Warning != expectedWarning {
		t.Errorf("Expected warning threshold %.1f, got %.1f", expectedWarning, thresholds.Warning)
	}
	if thresholds.Healthy != expectedHealthy {
		t.Errorf("Expected healthy threshold %.1f, got %.1f", expectedHealthy, thresholds.Healthy)
	}
}

func TestPhase10ERiskLevelMapping(t *testing.T) {
	tests := []struct {
		name     string
		rate     float64
		expected string
	}{
		{"Very low rate should be critical", 25.0, "critical"},
		{"Below critical threshold", 49.0, "critical"},
		{"At critical threshold", 50.0, "warning"},
		{"Between critical and warning", 65.0, "warning"},
		{"At warning threshold", 75.0, "acceptable"},
		{"Between warning and healthy", 85.0, "acceptable"},
		{"At healthy threshold", 90.0, "healthy"},
		{"Above healthy threshold", 95.0, "healthy"},
		{"Perfect rate", 100.0, "healthy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := receipts.GetContinuityRiskLevel(tt.rate)
			if result != tt.expected {
				t.Errorf("For rate %.1f, expected risk level '%s', got '%s'",
					tt.rate, tt.expected, result)
			}
		})
	}
}

func TestPhase10ERemediationItemStructure(t *testing.T) {
	item := receipts.RemediationItem{
		ReceiptID:    "receipt-123",
		EventID:      "event.domain.action",
		OldPolicyRef: "",
		NewPolicyRef: "policy.domain.action",
		Strategy:     "event-id-pattern",
		Success:      true,
		Error:        "",
	}

	// Test successful remediation item
	if !item.Success {
		t.Error("Expected success to be true")
	}
	if item.Error != "" {
		t.Error("Expected no error for successful remediation")
	}
	if item.Strategy == "" {
		t.Error("Expected strategy to be specified")
	}
	if item.NewPolicyRef == "" {
		t.Error("Expected new policy reference")
	}

	// Test failed remediation item
	failedItem := receipts.RemediationItem{
		ReceiptID:    "receipt-456",
		EventID:      "event.unknown.action",
		OldPolicyRef: "",
		NewPolicyRef: "",
		Strategy:     "domain-default",
		Success:      false,
		Error:        "policy validation failed",
	}

	if failedItem.Success {
		t.Error("Expected success to be false for failed item")
	}
	if failedItem.Error == "" {
		t.Error("Expected error message for failed remediation")
	}
}

func TestPhase10EContinuityBoundaryConditions(t *testing.T) {
	// Test edge cases for risk level calculation
	edgeCases := []struct {
		rate     float64
		expected string
	}{
		{0.0, "critical"},
		{49.999, "critical"},
		{50.0, "warning"},
		{74.999, "warning"},
		{75.0, "acceptable"},
		{89.999, "acceptable"},
		{90.0, "healthy"},
		{100.0, "healthy"},
	}

	for _, tc := range edgeCases {
		result := receipts.GetContinuityRiskLevel(tc.rate)
		if result != tc.expected {
			t.Errorf("Boundary case failed: rate %.3f expected '%s', got '%s'",
				tc.rate, tc.expected, result)
		}
	}
}
