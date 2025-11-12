package api_test

import (
	"net/http"
	"testing"

	"dis-core/internal/api"

	"github.com/go-chi/chi/v5"
)

// TestFormatVariantConsistency tests that API routes have proper format variants
func TestFormatVariantConsistency(t *testing.T) {
	// Create a test router with some sample routes
	router := chi.NewRouter()

	// Add some test routes that should have format variants
	router.Get("/api/domain/{id}/css", mockHandler)
	router.Get("/api/domain/{id}/css/json", mockHandler)
	router.Get("/api/domain/{id}/css/text", mockHandler)

	router.Get("/api/domain/{id}", mockHandler)
	router.Get("/api/domain/{id}/json", mockHandler)
	// Missing text variant - should be flagged

	router.Get("/api/policy/{id}", mockHandler)
	// Missing format variants - should be flagged

	// Test the format checker
	checker := api.NewFormatVariantChecker(router)

	// This should not panic
	checker.VerifyFormatConsistency()

	// In a real test, we would assert specific conditions
	// For now, this serves as a basic integration test
} // TestShouldHaveFormatVariants tests the logic for determining which routes need format support
func TestShouldHaveFormatVariants(t *testing.T) {
	router := chi.NewRouter()
	checker := api.NewFormatVariantChecker(router)

	testCases := []struct {
		path     string
		expected bool
		name     string
	}{
		{"/api/domain/123/css", true, "domain CSS should have format variants"},
		{"/api/domains", true, "domain list should have format variants"},
		{"/api/authority/status", true, "authority endpoints should have format variants"},
		{"/api/policy/abc", true, "policy endpoints should have format variants"},
		{"/api/receipts/verify/123", true, "receipt endpoints should have format variants"},
		{"/api/domain/123/files", false, "files endpoint typically doesn't need format variants"},
		{"/dashboard/receipts", false, "dashboard endpoints don't need format variants"},
		{"/internal/health", false, "non-API endpoints don't need format variants"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := checker.ShouldHaveFormatVariants(tc.path)
			if result != tc.expected {
				t.Errorf("ShouldHaveFormatVariants(%s) = %v, expected %v", tc.path, result, tc.expected)
			}
		})
	}
}

// TestParseRouteFormat tests the route format parsing logic
func TestParseRouteFormat(t *testing.T) {
	router := chi.NewRouter()
	checker := api.NewFormatVariantChecker(router)

	testCases := []struct {
		route        string
		expectedBase string
		expectedFmt  string
		name         string
	}{
		{"/api/domain/123/css/json", "/api/domain/123/css", "json", "explicit JSON format"},
		{"/api/domain/123/css/text", "/api/domain/123/css", "text", "explicit text format"},
		{"/api/domain/123", "/api/domain/123", "base", "base route without format"},
		{"/api/domain/123/{format}", "/api/domain/123", "format-param", "parameterized format"},
		{"/dashboard/receipts", "/dashboard/receipts", "base", "non-API route"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			base, format := checker.ParseRouteFormat(tc.route)
			if base != tc.expectedBase {
				t.Errorf("ParseRouteFormat(%s) base = %s, expected %s", tc.route, base, tc.expectedBase)
			}
			if format != tc.expectedFmt {
				t.Errorf("ParseRouteFormat(%s) format = %s, expected %s", tc.route, format, tc.expectedFmt)
			}
		})
	}
}

// mockHandler is a simple mock HTTP handler for testing
func mockHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("mock response"))
}
