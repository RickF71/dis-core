package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dis-core/internal/api"

	"github.com/go-chi/chi/v5"
)

// TestPhase10DEndpoints tests the Phase 10D Authority Console Schema Viewer integration
func TestPhase10DEndpoints(t *testing.T) {
	server := &api.Server{} // nil database for testing

	// Setup test router with Phase 10D routes
	router := chi.NewRouter()
	router.Get("/api/authority/schema/domain/{id}", server.GetAuthoritySchemaForDomain)
	router.Get("/api/authority/schema/domain/{id}/json", server.GetAuthoritySchemaForDomain)
	router.Get("/api/authority/schema/domain/{id}/text", server.GetAuthoritySchemaForDomain)
	// Allow the handler to receive other methods so it can return Method Not Allowed body
	router.Handle("/api/authority/schema/validate", http.HandlerFunc(server.HandleSchemaValidation))

	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		contentType    string
		expectedStatus int
		expectedBody   string
		checkJSON      bool
		checkYAML      bool
	}{
		{
			name:           "Domain Schema Aggregate JSON",
			method:         "GET",
			url:            "/api/authority/schema/domain/terra",
			expectedStatus: 200,
			checkJSON:      true,
		},
		{
			name:           "Domain Schema Aggregate Text/YAML",
			method:         "GET",
			url:            "/api/authority/schema/domain/terra/text",
			expectedStatus: 200,
			checkYAML:      true,
		},
		{
			name:           "Domain Schema Aggregate Auto-Detection JSON",
			method:         "GET",
			url:            "/api/authority/schema/domain/terra",
			expectedStatus: 200,
			checkJSON:      true,
		},
		{
			name:           "Empty Domain Schema Aggregate",
			method:         "GET",
			url:            "/api/authority/schema/domain/empty",
			expectedStatus: 200,
			checkJSON:      true,
		},
		{
			name:           "Missing Domain ID",
			method:         "GET",
			url:            "/api/authority/schema/domain/",
			expectedStatus: 404,
		},
		{
			name:        "Valid Schema Validation",
			method:      "POST",
			url:         "/api/authority/schema/validate",
			contentType: "application/json",
			body: `{
				"schema": {
					"type": "object",
					"$schema": "http://json-schema.org/draft-07/schema#",
					"properties": {
						"name": {"type": "string"}
					},
					"required": ["name"]
				}
			}`,
			expectedStatus: 200,
			checkJSON:      true,
		},
		{
			name:           "Invalid Schema Validation - Empty",
			method:         "POST",
			url:            "/api/authority/schema/validate",
			contentType:    "application/json",
			body:           `{"schema": {}}`,
			expectedStatus: 200,
			expectedBody:   "Schema cannot be empty",
		},
		{
			name:           "Invalid Schema Validation - Wrong Type",
			method:         "POST",
			url:            "/api/authority/schema/validate",
			contentType:    "application/json",
			body:           `{"schema": "not an object"}`,
			expectedStatus: 200,
			expectedBody:   "Schema must be a JSON object",
		},
		{
			name:        "Schema Validation YAML Input",
			method:      "POST",
			url:         "/api/authority/schema/validate",
			contentType: "text/yaml",
			body: `type: object
properties:
  name:
    type: string
required:
  - name`,
			expectedStatus: 200,
			checkJSON:      true,
		},
		{
			name:           "Invalid JSON Body",
			method:         "POST",
			url:            "/api/authority/schema/validate",
			contentType:    "application/json",
			body:           `{"invalid": json}`,
			expectedStatus: 400,
			expectedBody:   "Invalid JSON format",
		},
		{
			name:           "Method Not Allowed for Validation",
			method:         "GET",
			url:            "/api/authority/schema/validate",
			expectedStatus: 405,
			expectedBody:   "Method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
				if tt.contentType != "" {
					req.Header.Set("Content-Type", tt.contentType)
				}
			} else {
				req = httptest.NewRequest(tt.method, tt.url, nil)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check expected body content
			if tt.expectedBody != "" {
				if !strings.Contains(rr.Body.String(), tt.expectedBody) {
					t.Errorf("Expected body to contain %q, got %q", tt.expectedBody, rr.Body.String())
				}
			}

			// Check JSON structure
			if tt.checkJSON && rr.Code == 200 {
				var jsonResponse map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &jsonResponse); err != nil {
					t.Errorf("Expected valid JSON response, got error: %v", err)
				}

				// Validate specific fields for different endpoints
				if strings.Contains(tt.url, "/domain/") {
					// Domain aggregate response
					if _, exists := jsonResponse["domain_ref"]; !exists {
						t.Error("Expected domain_ref field in aggregate response")
					}
					if _, exists := jsonResponse["timestamp"]; !exists {
						t.Error("Expected timestamp field in aggregate response")
					}
					if _, exists := jsonResponse["schema"]; !exists {
						t.Error("Expected schema field in aggregate response")
					}
					if _, exists := jsonResponse["policies"]; !exists {
						t.Error("Expected policies field in aggregate response")
					}
					if _, exists := jsonResponse["policy_count"]; !exists {
						t.Error("Expected policy_count field in aggregate response")
					}
				} else if strings.Contains(tt.url, "/validate") {
					// Validation response
					if _, exists := jsonResponse["valid"]; !exists {
						t.Error("Expected valid field in validation response")
					}
					if _, exists := jsonResponse["errors"]; !exists {
						t.Error("Expected errors field in validation response")
					}
					if _, exists := jsonResponse["timestamp"]; !exists {
						t.Error("Expected timestamp field in validation response")
					}
					if _, exists := jsonResponse["validated_by"]; !exists {
						t.Error("Expected validated_by field in validation response")
					}
				}
			}

			// Check YAML structure
			if tt.checkYAML && rr.Code == 200 {
				body := rr.Body.String()
				if !strings.Contains(body, "domain_ref:") {
					t.Error("Expected YAML response to contain domain_ref field")
				}
				if !strings.Contains(body, "timestamp:") {
					t.Error("Expected YAML response to contain timestamp field")
				}
				if !strings.Contains(body, "schema:") {
					t.Error("Expected YAML response to contain schema field")
				}
			}

			// Check content type headers
			if tt.checkJSON {
				expectedContentType := "application/json"
				if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, expectedContentType) {
					t.Errorf("Expected Content-Type to contain %s, got %s", expectedContentType, contentType)
				}
			}
			if tt.checkYAML {
				expectedContentType := "text/plain"
				if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, expectedContentType) {
					t.Errorf("Expected Content-Type to contain %s, got %s", expectedContentType, contentType)
				}
			}
		})
	}
}

// TestPhase10DIntegration tests the integration between Phase 10C and 10D endpoints
func TestPhase10DIntegration(t *testing.T) {
	t.Run("Integration Validation", func(t *testing.T) {
		server := &api.Server{} // nil database for testing

		// Test that Phase 10D endpoints integrate with Phase 10C logic
		router := chi.NewRouter()
		router.Get("/api/authority/schema/domain/{id}", server.GetAuthoritySchemaForDomain)

		req := httptest.NewRequest("GET", "/api/authority/schema/domain/test", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Should return 200 with aggregated data structure
		if rr.Code != 200 {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Errorf("Expected valid JSON response: %v", err)
		}

		// Verify integration structure
		if domainRef, ok := response["domain_ref"].(string); !ok || domainRef != "test" {
			t.Error("Expected domain_ref to match requested domain ID")
		}

		t.Logf("Phase 10D integration test successful")
	})
}

// TestPhase10DSchemaValidationCoverage tests comprehensive schema validation scenarios
func TestPhase10DSchemaValidationCoverage(t *testing.T) {
	t.Run("Schema Validation Coverage", func(t *testing.T) {
		server := &api.Server{}

		// Test various schema validation scenarios
		validationTests := []struct {
			name     string
			schema   interface{}
			expected bool
			errors   []string
		}{
			{
				name: "Valid Terra Domain Schema",
				schema: map[string]interface{}{
					"type":    "object",
					"$schema": "http://json-schema.org/draft-07/schema#",
					"title":   "Terra Domain Schema",
					"properties": map[string]interface{}{
						"geography": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"coordinates": map[string]interface{}{"type": "array"},
							},
						},
					},
					"required": []interface{}{"geography"},
				},
				expected: true,
				errors:   []string{},
			},
			{
				name:     "Empty Schema",
				schema:   map[string]interface{}{},
				expected: false,
				errors:   []string{"Schema cannot be empty"},
			},
			{
				name: "Missing Properties",
				schema: map[string]interface{}{
					"type":  "object",
					"title": "Test Schema",
				},
				expected: false,
				errors:   []string{"Schema should contain 'properties' field"},
			},
		}

		for _, vt := range validationTests {
			t.Run(vt.name, func(t *testing.T) {
				result := server.ValidateSchemaStructure(vt.schema)

				if result.Valid != vt.expected {
					t.Errorf("Expected validation result %v, got %v", vt.expected, result.Valid)
				}

				// Check that expected errors are present
				for _, expectedError := range vt.errors {
					found := false
					for _, actualError := range result.Errors {
						if actualError == expectedError {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error %q not found in result errors: %v", expectedError, result.Errors)
					}
				}

				// Verify response structure
				if result.Timestamp == "" {
					t.Error("Expected timestamp to be set")
				}
				if result.ValidatedBy == "" {
					t.Error("Expected validated_by to be set")
				}

				t.Logf("Validation test %q passed", vt.name)
			})
		}
	})
}
