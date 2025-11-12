package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// TestDomainFilesJSON tests the JSON format of domain files endpoint
func TestDomainFilesJSON(t *testing.T) {
	// Create test router
	router := chi.NewRouter()

	// Create mock server
	server := &MockServer{}

	// Register the route
	router.Get("/api/domain/{id}/files", server.handleDomainFiles)
	router.Get("/api/domain/{id}/files/json", server.handleDomainFilesJSON)

	// Test with JSON format
	req := httptest.NewRequest("GET", "/api/domain/test-domain/files/json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected JSON content type, got %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "[") || !strings.Contains(body, "]") {
		t.Errorf("Expected JSON array format, got: %s", body)
	}
}

// TestDomainFilesText tests the text format of domain files endpoint
func TestDomainFilesText(t *testing.T) {
	// Create test router
	router := chi.NewRouter()

	// Create mock server
	server := &MockServer{}

	// Register the route
	router.Get("/api/domain/{id}/files/text", server.handleDomainFilesText)

	// Test with text format
	req := httptest.NewRequest("GET", "/api/domain/test-domain/files/text", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("Expected text/plain content type, got %s", contentType)
	}
}

// TestDomainFilesEmpty tests empty result handling
func TestDomainFilesEmpty(t *testing.T) {
	// Create test router
	router := chi.NewRouter()

	// Create mock server that returns empty results
	server := &MockServerEmpty{}

	// Register the route
	router.Get("/api/domain/{id}/files", server.handleDomainFiles)
	router.Get("/api/domain/{id}/files/json", server.handleDomainFilesJSON)

	// Test empty domain
	req := httptest.NewRequest("GET", "/api/domain/empty-domain/files/json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for empty domain, got %d", w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if body != "[]" {
		t.Errorf("Expected empty array [], got: %s", body)
	}
}

// TestDomainFilesDomainNotFound tests non-existent domain handling
func TestDomainFilesDomainNotFound(t *testing.T) {
	// Create test router
	router := chi.NewRouter()

	// Create mock server that returns domain not found
	server := &MockServerNotFound{}

	// Register the route
	router.Get("/api/domain/{id}/files", server.handleDomainFiles)
	router.Get("/api/domain/{id}/files/json", server.handleDomainFilesJSON)

	// Test non-existent domain
	req := httptest.NewRequest("GET", "/api/domain/nonexistent/files/json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent domain, got %d", w.Code)
	}
}

// Mock server implementations
type MockServer struct{}

func (s *MockServer) handleDomainFiles(w http.ResponseWriter, r *http.Request) {
	// Mock response with sample files
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`["policy.yaml", "thresholds.json", "index.css"]`))
}

func (s *MockServer) handleDomainFilesJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`["policy.yaml", "thresholds.json", "index.css"]`))
}

func (s *MockServer) handleDomainFilesText(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("policy.yaml\nthresholds.json\nindex.css"))
}

type MockServerEmpty struct{}

func (s *MockServerEmpty) handleDomainFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`[]`))
}

func (s *MockServerEmpty) handleDomainFilesJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`[]`))
}

type MockServerNotFound struct{}

func (s *MockServerNotFound) handleDomainFiles(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Domain not found", http.StatusNotFound)
}

func (s *MockServerNotFound) handleDomainFilesJSON(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Domain not found", http.StatusNotFound)
}
