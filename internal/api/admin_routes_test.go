package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestVerifySeatRole tests the role verification logic directly
func TestVerifySeatRole(t *testing.T) {
	server := &Server{}

	tests := []struct {
		authHeader   string
		expectedRole string
		expectError  bool
		description  string
	}{
		{"", "", true, "Empty header should error"},
		{"Bearer admin-root-token", "root", false, "Admin root token should return root role"},
		{"Bearer policy-admin-token", "root", false, "Policy admin token should return root role"},
		{"Bearer some-admin-token", "policy.admin", false, "Admin-containing token should return policy.admin"},
		{"Bearer invalid-token", "", true, "Invalid token should error"},
		{"Invalid format", "", true, "Non-Bearer format should error"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			role, err := server.verifySeatRole(req)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && role != tt.expectedRole {
				t.Errorf("Expected role '%s', got '%s'", tt.expectedRole, role)
			}
		})
	}
}

// TestAdminRoutes_Authorization tests the admin token verification
func TestAdminRoutes_Authorization(t *testing.T) {
	server := &Server{
		mux: http.NewServeMux(),
	}
	server.registerAdminRoutes()

	tests := []struct {
		name         string
		authHeader   string
		expectedCode int
		description  string
	}{
		{"No auth header", "", http.StatusUnauthorized, "Should reject requests without authorization"},
		{"Invalid format", "Invalid token", http.StatusUnauthorized, "Should reject non-Bearer tokens"},
		{"Invalid token", "Bearer invalid-token", http.StatusUnauthorized, "Should reject invalid tokens"},
		// Note: Valid tokens will fail with 500 due to nil DB, but that means auth passed
		{"Valid admin token", "Bearer admin-root-token", http.StatusInternalServerError, "Should accept admin tokens"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/admin/receipts/recent", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			server.mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("%s: Expected status %d, got %d. %s", tt.name, tt.expectedCode, rr.Code, tt.description)
			}
		})
	}
}

// TestAdminRoutes_FreezeRequest tests the freeze endpoint request handling
func TestAdminRoutes_FreezeRequest(t *testing.T) {
	server := &Server{
		mux:    http.NewServeMux(),
		ledger: nil, // Will succeed up to ledger recording
	}
	server.registerAdminRoutes()

	// Test valid freeze request
	freezeReq := AdminFreezeRequest{
		Target: "domain.test",
		Reason: "Testing freeze functionality",
	}
	body, _ := json.Marshal(freezeReq)

	req := httptest.NewRequest("POST", "/api/admin/freeze", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer admin-root-token")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Response: %s", rr.Code, rr.Body.String())
	}

	// Verify response contains expected fields
	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if status, ok := response["status"].(string); !ok || status != "success" {
		t.Errorf("Expected status 'success', got %v", response["status"])
	}

	if target, ok := response["target"].(string); !ok || target != "domain.test" {
		t.Errorf("Expected target 'domain.test', got %v", response["target"])
	}
}

// TestAdminRoutes_InvalidFreezeRequest tests invalid freeze requests
func TestAdminRoutes_InvalidFreezeRequest(t *testing.T) {
	server := &Server{
		mux:    http.NewServeMux(),
		ledger: nil,
	}
	server.registerAdminRoutes()

	// Test empty target
	emptyTargetReq := AdminFreezeRequest{Target: "", Reason: "Testing"}
	body, _ := json.Marshal(emptyTargetReq)

	req := httptest.NewRequest("POST", "/api/admin/freeze", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer admin-root-token")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty target, got %d", rr.Code)
	}
}
