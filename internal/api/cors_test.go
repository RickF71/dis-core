package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dis-core/internal/cors"

	"github.com/go-chi/chi/v5"
)

// Required CORS headers for MOAR-CORS v1
var requiredHeaders = []string{
	"Access-Control-Allow-Origin",
	"Access-Control-Allow-Credentials",
	"Access-Control-Allow-Headers",
	"Access-Control-Allow-Methods",
}

// getAllRoutes extracts all routes from the chi router by walking it
func getAllRoutes(router *chi.Mux) []string {
	routes := []string{}

	chi.Walk(router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		// Collect all GET/POST routes (they should support OPTIONS preflight)
		if method == "GET" || method == "POST" {
			routes = append(routes, route)
		}
		return nil
	})

	return routes
}

func TestAllRoutesHaveCORS(t *testing.T) {
	// Create a test server instance
	server := &Server{
		router: NewFormatAwareRouter(),
		db:     nil, // No DB needed for CORS tests
	}

	// Apply CORS middleware (CRITICAL for tests)
	server.router.Use(cors.Middleware)

	// Register routes (minimal setup for testing)
	server.RegisterAllRoutes()

	routes := getAllRoutes(server.router)

	if len(routes) == 0 {
		t.Fatalf("no routes discovered; chi.Walk may not be functioning")
	}

	t.Logf("Testing CORS on %d routes", len(routes))

	for _, route := range routes {
		// Replace chi route params with dummy values
		testRoute := strings.ReplaceAll(route, "{id}", "test-id")
		testRoute = strings.ReplaceAll(testRoute, "{domain_id}", "test-domain-id")
		testRoute = strings.ReplaceAll(testRoute, "{challenge_id}", "test-challenge-id")
		testRoute = strings.ReplaceAll(testRoute, "{alias}", "test-alias")
		testRoute = strings.ReplaceAll(testRoute, "{seat_id}", "test-seat-id")
		testRoute = strings.ReplaceAll(testRoute, "{contract_id}", "test-contract-id")

		req := httptest.NewRequest("OPTIONS", testRoute, nil)
		req.Header.Set("Origin", "http://localhost:5173") // Simulate browser preflight
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)
		resp := w.Result()

		// OPTIONS should return 200 or 204
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
			t.Errorf("route %s returned unexpected status %d for OPTIONS", route, resp.StatusCode)
			continue
		}

		// Check all required CORS headers
		for _, h := range requiredHeaders {
			if resp.Header.Get(h) == "" {
				t.Errorf("route %s missing required CORS header: %s", route, h)
			}
		}

		// Validate Access-Control-Allow-Origin
		allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
		if allowOrigin == "" {
			t.Errorf("route %s missing Access-Control-Allow-Origin", route)
			continue
		}

		// Should be specific origin (not wildcard) for credential safety
		if allowOrigin == "*" {
			t.Errorf("route %s uses wildcard origin (INSECURE with credentials)", route)
		}

		// Should match request origin (dynamic CORS)
		if !strings.Contains(allowOrigin, "localhost") && !strings.Contains(allowOrigin, "127.0.0.1") {
			t.Errorf("route %s returned invalid Access-Control-Allow-Origin: %s", route, allowOrigin)
		}

		// Validate Access-Control-Allow-Credentials
		allowCreds := resp.Header.Get("Access-Control-Allow-Credentials")
		if allowCreds != "true" {
			t.Errorf("route %s missing or invalid Access-Control-Allow-Credentials: %s", route, allowCreds)
		}
	}
}

func TestCORSWithDisallowedOrigin(t *testing.T) {
	server := &Server{
		router: NewFormatAwareRouter(),
		db:     nil,
	}
	server.router.Use(cors.Middleware)
	server.RegisterAllRoutes()

	req := httptest.NewRequest("OPTIONS", "/api/status", nil)
	req.Header.Set("Origin", "http://evil.com") // Disallowed origin
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)
	resp := w.Result()

	// Should still return 200 OK but without CORS headers
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for disallowed origin, got %d", resp.StatusCode)
	}

	// Should NOT include CORS headers for disallowed origins
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if allowOrigin != "" {
		t.Errorf("Disallowed origin should not receive CORS headers, got: %s", allowOrigin)
	}
}

func TestCORSWithAllowedOrigins(t *testing.T) {
	server := &Server{
		router: NewFormatAwareRouter(),
		db:     nil,
	}
	server.router.Use(cors.Middleware)
	server.RegisterAllRoutes()

	allowedOrigins := []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://[::1]:5173",
	}

	for _, origin := range allowedOrigins {
		req := httptest.NewRequest("OPTIONS", "/api/status", nil)
		req.Header.Set("Origin", origin)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)
		resp := w.Result()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 OK for allowed origin %s, got %d", origin, resp.StatusCode)
		}

		allowOriginHeader := resp.Header.Get("Access-Control-Allow-Origin")
		if allowOriginHeader != origin {
			t.Errorf("Expected Access-Control-Allow-Origin: %s, got: %s", origin, allowOriginHeader)
		}

		allowCreds := resp.Header.Get("Access-Control-Allow-Credentials")
		if allowCreds != "true" {
			t.Errorf("Expected Access-Control-Allow-Credentials: true, got: %s", allowCreds)
		}
	}
}

func TestCORSHeadersComplete(t *testing.T) {
	server := &Server{
		router: NewFormatAwareRouter(),
		db:     nil,
	}
	server.router.Use(cors.Middleware)
	server.RegisterAllRoutes()

	req := httptest.NewRequest("OPTIONS", "/api/status", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)
	resp := w.Result()

	// Validate all expected headers
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":      "http://localhost:5173",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Methods":     "GET,POST,OPTIONS",
		"Vary":                             "Origin",
	}

	for header, expectedValue := range expectedHeaders {
		actualValue := resp.Header.Get(header)
		if actualValue == "" {
			t.Errorf("Missing header: %s", header)
		} else if header != "Access-Control-Allow-Headers" && actualValue != expectedValue {
			t.Errorf("Header %s: expected %s, got %s", header, expectedValue, actualValue)
		}
	}

	// Validate Access-Control-Allow-Headers includes required headers
	allowHeaders := resp.Header.Get("Access-Control-Allow-Headers")
	requiredAllowHeaders := []string{
		"Content-Type",
		"Authorization",
		"X-External-User",
		"X-Acting-As",
		"X-Requested-With",
		"Accept",
	}

	for _, h := range requiredAllowHeaders {
		if !strings.Contains(allowHeaders, h) {
			t.Errorf("Access-Control-Allow-Headers missing: %s", h)
		}
	}
}

func TestCORSNoOriginHeader(t *testing.T) {
	server := &Server{
		router: NewFormatAwareRouter(),
		db:     nil,
	}
	server.router.Use(cors.Middleware)
	server.RegisterAllRoutes()

	// Request without Origin header (same-origin request)
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)
	resp := w.Result()

	// Should work but without CORS headers
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK for same-origin request, got %d", resp.StatusCode)
	}

	// CORS headers may or may not be present for same-origin requests
	// (middleware may skip or still add them - both are acceptable)
	t.Logf("Same-origin request CORS headers: %v", resp.Header.Get("Access-Control-Allow-Origin"))
}
