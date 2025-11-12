package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dis-core/internal/receipts"
)

// MockServer represents a test server with mocked dependencies
type MockServer struct {
	router chi.Router
}

// setupMockServer creates a test server for API testing
func setupMockServer() *MockServer {
	r := chi.NewRouter()

	// Mock handlers that simulate Phase 10F endpoints without database
	r.Post("/api/receipts/fix", mockCreateFixReceiptHandler)
	r.Get("/api/receipts/fix", mockListFixReceiptsHandler)
	r.Get("/api/receipts/lineage/{id}", mockGetLineageProofHandler)

	return &MockServer{router: r}
}

// TestFixReceiptAPIEndpoints tests the Phase 10F API endpoints
func TestFixReceiptAPIEndpoints(t *testing.T) {
	server := setupMockServer()

	t.Run("POST /api/receipts/fix - Create fix receipt", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"original_receipt": "550e8400-e29b-41d4-a716-446655440000",
			"fix_method":       "pattern-match",
			"domain_ref":       "domain.test.v1",
			"authorized_by":    "authority.test",
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/receipts/fix", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response receipts.FixReceipt
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, strings.HasPrefix(response.ID, "rcpt-fix-"))
		assert.Equal(t, "pattern-match", response.FixMethod)
		assert.Equal(t, "domain.test.v1", response.DomainRef)
	})

	t.Run("GET /api/receipts/fix - List fix receipts", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/receipts/fix", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response []receipts.FixReceipt
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Mock should return at least one fix receipt
		assert.GreaterOrEqual(t, len(response), 1)
	})

	t.Run("GET /api/receipts/lineage/{id} - Get lineage proof", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/receipts/lineage/rcpt-fix-test-001", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response receipts.LineageProof
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "rcpt-fix-test-001", response.FixReceiptID)
		assert.NotEmpty(t, response.OriginalReceiptID)
	})
}

// TestAPIFormatHandling tests JSON vs text format responses
func TestAPIFormatHandling(t *testing.T) {
	server := setupMockServer()

	t.Run("JSON format response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/receipts/fix", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	})
}

// Mock handlers for testing without database dependencies

func mockCreateFixReceiptHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OriginalReceipt string `json:"original_receipt"`
		FixMethod       string `json:"fix_method"`
		DomainRef       string `json:"domain_ref"`
		AuthorizedBy    string `json:"authorized_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create mock fix receipt
	fixReceipt := receipts.FixReceipt{
		ID:              "rcpt-fix-mock-001",
		OriginalReceipt: req.OriginalReceipt,
		FixMethod:       req.FixMethod,
		DomainRef:       req.DomainRef,
		AuthorizedBy:    req.AuthorizedBy,
		Timestamp:       time.Now().UTC(),
		Verification:    receipts.VerificationPending,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fixReceipt)
}

func mockListFixReceiptsHandler(w http.ResponseWriter, r *http.Request) {
	// Mock fix receipts list
	fixReceipts := []receipts.FixReceipt{
		{
			ID:              "rcpt-fix-mock-001",
			OriginalReceipt: "550e8400-e29b-41d4-a716-446655440000",
			FixMethod:       receipts.FixMethodPatternMatch,
			DomainRef:       "domain.test.v1",
			AuthorizedBy:    "authority.test",
			Timestamp:       time.Now().UTC(),
			Verification:    receipts.VerificationHash,
		},
		{
			ID:              "rcpt-fix-mock-002",
			OriginalReceipt: "550e8400-e29b-41d4-a716-446655440001",
			FixMethod:       receipts.FixMethodDomainDefault,
			DomainRef:       "domain.test.v2",
			AuthorizedBy:    "authority.test",
			Timestamp:       time.Now().UTC(),
			Verification:    receipts.VerificationPending,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fixReceipts)
}

func mockGetLineageProofHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Create mock lineage proof
	lineageProof := receipts.LineageProof{
		OriginalReceiptID: "550e8400-e29b-41d4-a716-446655440000",
		FixReceiptID:      id,
		ProofChain:        []string{"proof-001", "proof-002"},
		Verified:          true,
		CreatedAt:         time.Now().UTC(),
		VerifiedAt:        func() *time.Time { t := time.Now().UTC(); return &t }(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lineageProof)
}
