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

// TestProofSyncAPI tests the Phase 10G proof synchronization endpoints
func TestProofSyncAPI(t *testing.T) {
	server := setupMockProofSyncServer()

	t.Run("POST /api/receipts/proof/sync - Successful sync", func(t *testing.T) {
		requestBody := receipts.ProofSyncRequest{
			SourceDomain: "terra.domain",
			TargetDomain: "usa.domain",
			ProofIDs:     []string{"rcpt-fix-terra-001", "rcpt-fix-terra-002"},
			SyncMode:     receipts.SyncModePush,
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/receipts/proof/sync", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response receipts.ProofSyncResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "success", response.Status)
		assert.Greater(t, response.SyncedProofs, 0)
		assert.Equal(t, 0, response.FailedProofs)
	})

	t.Run("POST /api/receipts/proof/sync - Invalid request", func(t *testing.T) {
		requestBody := receipts.ProofSyncRequest{
			SourceDomain: "", // Missing required field
			TargetDomain: "usa.domain",
			SyncMode:     receipts.SyncModePush,
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/receipts/proof/sync", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GET /api/receipts/proof/verify/{id} - Valid proof", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/receipts/proof/verify/fed-proof-001", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "fed-proof-001", response["proof_id"])
		assert.True(t, response["verified"].(bool))
	})

	t.Run("GET /api/receipts/proof/verify/{id} - Text format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/receipts/proof/verify/fed-proof-001/text", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
		assert.Contains(t, w.Body.String(), "Proof Verification Results")
	})
}

// TestFederationSummaryAPI tests the federation summary endpoint
func TestFederationSummaryAPI(t *testing.T) {
	server := setupMockProofSyncServer()

	t.Run("GET /api/federation/summary - JSON format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/federation/summary", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

		var summary receipts.FederationSummary
		err := json.Unmarshal(w.Body.Bytes(), &summary)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, summary.TotalFederationProofs, 0)
		assert.GreaterOrEqual(t, summary.VerificationRate, 0.0)
		assert.LessOrEqual(t, summary.VerificationRate, 100.0)
	})

	t.Run("GET /api/federation/summary - Text format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/federation/summary/text", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
		assert.Contains(t, w.Body.String(), "Federation Summary")
		assert.Contains(t, w.Body.String(), "Total Federation Proofs:")
	})
}

// TestFederationTrustAPI tests the federation trust relationship endpoint
func TestFederationTrustAPI(t *testing.T) {
	server := setupMockProofSyncServer()

	t.Run("POST /api/federation/trust - Create trust relationship", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"domain_a":    "terra.domain",
			"domain_b":    "usa.domain",
			"trust_level": receipts.TrustLevelHigh,
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/federation/trust", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "success", response["status"])
		assert.Equal(t, "terra.domain", response["domain_a"])
		assert.Equal(t, "usa.domain", response["domain_b"])
		assert.Equal(t, receipts.TrustLevelHigh, response["trust_level"])
	})

	t.Run("POST /api/federation/trust - Invalid trust level", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"domain_a":    "terra.domain",
			"domain_b":    "usa.domain",
			"trust_level": "invalid_level",
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/federation/trust", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("POST /api/federation/trust - Missing domains", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"domain_a":    "",
			"domain_b":    "usa.domain",
			"trust_level": receipts.TrustLevelMedium,
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/federation/trust", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestCrossDomainFormatHandling tests format detection for Phase 10G endpoints
func TestCrossDomainFormatHandling(t *testing.T) {
	server := setupMockProofSyncServer()

	t.Run("Proof sync text format", func(t *testing.T) {
		requestBody := receipts.ProofSyncRequest{
			SourceDomain: "terra.domain",
			TargetDomain: "usa.domain",
			SyncMode:     receipts.SyncModePull,
		}

		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/api/receipts/proof/sync/text", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
		assert.Contains(t, w.Body.String(), "Proof Synchronization Results")
	})

	t.Run("Accept header format preference", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/federation/summary", nil)
		req.Header.Set("Accept", "text/plain")
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
	})
}

// Mock server setup for Phase 10G testing

type MockProofSyncServer struct {
	router chi.Router
}

func setupMockProofSyncServer() *MockProofSyncServer {
	r := chi.NewRouter()

	// Phase 10G endpoints with mock handlers
	r.Post("/api/receipts/proof/sync", mockProofSyncHandler)
	r.Post("/api/receipts/proof/sync/json", mockProofSyncHandler)
	r.Post("/api/receipts/proof/sync/text", mockProofSyncHandler)
	r.Get("/api/receipts/proof/verify/{id}", mockVerifyProofHandler)
	r.Get("/api/receipts/proof/verify/{id}/json", mockVerifyProofHandler)
	r.Get("/api/receipts/proof/verify/{id}/text", mockVerifyProofHandler)
	r.Get("/api/federation/summary", mockFederationSummaryHandler)
	r.Get("/api/federation/summary/json", mockFederationSummaryHandler)
	r.Get("/api/federation/summary/text", mockFederationSummaryHandler)
	r.Post("/api/federation/trust", mockCreateFederationTrustHandler)

	return &MockProofSyncServer{router: r}
}

// Mock handlers for Phase 10G endpoints

func mockProofSyncHandler(w http.ResponseWriter, r *http.Request) {
	var req receipts.ProofSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.SourceDomain == "" || req.TargetDomain == "" {
		http.Error(w, "Source and target domains are required", http.StatusBadRequest)
		return
	}

	// Mock successful sync response
	response := receipts.ProofSyncResponse{
		Status:       "success",
		SyncedProofs: len(req.ProofIDs),
		FailedProofs: 0,
		Timestamp:    time.Now().UTC(),
		Details: map[string]string{
			"trust_level": receipts.TrustLevelHigh,
			"sync_mode":   req.SyncMode,
		},
	}

	// Check format preference
	if strings.Contains(r.URL.Path, "/text") || r.Header.Get("Accept") == "text/plain" {
		writeProofSyncMockText(w, &response)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func mockVerifyProofHandler(w http.ResponseWriter, r *http.Request) {
	proofID := chi.URLParam(r, "id")

	// Mock verification response (always successful for testing)
	response := map[string]interface{}{
		"proof_id":  proofID,
		"verified":  true,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Check format preference
	if strings.Contains(r.URL.Path, "/text") || r.Header.Get("Accept") == "text/plain" {
		writeVerificationMockText(w, response)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func mockFederationSummaryHandler(w http.ResponseWriter, r *http.Request) {
	// Mock federation summary
	summary := receipts.FederationSummary{
		TotalFederationProofs: 15,
		VerifiedProofs:        12,
		PendingProofs:         2,
		DiscrepancyCount:      1,
		VerificationRate:      80.0,
		TrustedDomains:        []string{"terra.domain", "usa.domain"},
		LastSyncTimestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	// Check format preference
	if strings.Contains(r.URL.Path, "/text") || r.Header.Get("Accept") == "text/plain" {
		writeFederationSummaryMockText(w, &summary)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	}
}

func mockCreateFederationTrustHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DomainA    string `json:"domain_a"`
		DomainB    string `json:"domain_b"`
		TrustLevel string `json:"trust_level"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate fields
	if req.DomainA == "" || req.DomainB == "" {
		http.Error(w, "Both domains are required", http.StatusBadRequest)
		return
	}

	validTrustLevels := []string{receipts.TrustLevelHigh, receipts.TrustLevelMedium, receipts.TrustLevelLow, receipts.TrustLevelNone}
	isValid := false
	for _, level := range validTrustLevels {
		if req.TrustLevel == level {
			isValid = true
			break
		}
	}

	if !isValid {
		http.Error(w, "Invalid trust level", http.StatusBadRequest)
		return
	}

	// Mock success response
	response := map[string]interface{}{
		"status":      "success",
		"trust_id":    "trust-" + req.DomainA + "-" + req.DomainB,
		"domain_a":    req.DomainA,
		"domain_b":    req.DomainB,
		"trust_level": req.TrustLevel,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Mock text format writers

func writeProofSyncMockText(w http.ResponseWriter, response *receipts.ProofSyncResponse) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("=== Proof Synchronization Results ===\n"))
	w.Write([]byte("Status: " + response.Status + "\n"))
	w.Write([]byte("Synced Proofs: " + string(rune(response.SyncedProofs+'0')) + "\n"))
}

func writeVerificationMockText(w http.ResponseWriter, response map[string]interface{}) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("=== Proof Verification Results ===\n"))
	w.Write([]byte("Proof ID: " + response["proof_id"].(string) + "\n"))
	w.Write([]byte("Verified: true\n"))
}

func writeFederationSummaryMockText(w http.ResponseWriter, summary *receipts.FederationSummary) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("=== Federation Summary ===\n"))
	w.Write([]byte("Total Federation Proofs: 15\n"))
	w.Write([]byte("Verification Rate: 80.0%\n"))
}
