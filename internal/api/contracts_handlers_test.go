package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dis-core/internal/contracts"
	"dis-core/internal/identity"
	"dis-core/internal/repo"
	"dis-core/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GOV-13: Contract API Handler Tests

// setupContractAPITestDB creates a test database connection
func setupContractAPITestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	dsn := "postgres://dis_user:card567@localhost:5432/dis?sslmode=disable"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("Skipping API test: cannot connect to test database: %v", err)
	}
	return pool
}

// setupContractAPITestDomains creates test domains
func setupContractAPITestDomains(t *testing.T, pool *pgxpool.Pool) (issuerID, subjectID uuid.UUID) {
	ctx := context.Background()

	issuerID = uuid.New()
	subjectID = uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO domains (id, name, created_at)
		VALUES
			($1, 'api-test-issuer', NOW()),
			($2, 'api-test-subject', NOW())
		ON CONFLICT (id) DO NOTHING
	`, issuerID, subjectID)

	if err != nil {
		t.Fatalf("Failed to create test domains: %v", err)
	}

	return issuerID, subjectID
}

// cleanupContractAPIData removes test data
func cleanupContractAPIData(t *testing.T, pool *pgxpool.Pool, domainIDs ...uuid.UUID) {
	ctx := context.Background()
	for _, id := range domainIDs {
		_, _ = pool.Exec(ctx, "DELETE FROM contracts WHERE domain_id = $1 OR subject_domain_id = $1", id)
		_, _ = pool.Exec(ctx, "DELETE FROM domains WHERE id = $1", id)
	}
}

func TestHandleCreateContract_Success(t *testing.T) {
	pool := setupContractAPITestDB(t)
	defer pool.Close()

	issuerID, subjectID := setupContractAPITestDomains(t, pool)
	defer cleanupContractAPIData(t, pool, issuerID, subjectID)

	// Create test server with contract service
	contractRepo := repo.NewContractRepository(pool)
	receiptStore := identity.NewIdentityReceiptStore(pool)
	contractService := services.NewContractService(contractRepo, receiptStore)

	server := &Server{
		db:              pool,
		router:          chi.NewRouter(),
		contractService: contractService,
	}

	// Build request payload
	effectiveAt := time.Now().UTC().Add(-1 * time.Hour)
	payload := CreateContractRequest{
		SubjectDomainID: subjectID.String(),
		ContractType:    "membership",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/membership/v1",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     effectiveAt.Format(time.RFC3339),
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/domain/"+issuerID.String()+"/contracts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", issuerID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Execute request
	w := httptest.NewRecorder()
	server.HandleCreateContract(w, req)

	// Assert 201 Created
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	// Decode response
	var response ContractDTO
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Assert fields
	if response.DomainID != issuerID {
		t.Errorf("Expected domain_id %s, got %s", issuerID, response.DomainID)
	}
	if response.SubjectDomainID != subjectID {
		t.Errorf("Expected subject_domain_id %s, got %s", subjectID, response.SubjectDomainID)
	}
	if response.ContractType != "membership" {
		t.Errorf("Expected contract_type 'membership', got %s", response.ContractType)
	}
	if response.Status != "active" {
		t.Errorf("Expected status 'active', got %s", response.Status)
	}
}

func TestHandleCreateContract_ValidationErrors(t *testing.T) {
	pool := setupContractAPITestDB(t)
	defer pool.Close()

	issuerID, subjectID := setupContractAPITestDomains(t, pool)
	defer cleanupContractAPIData(t, pool, issuerID, subjectID)

	contractRepo := repo.NewContractRepository(pool)
	receiptStore := identity.NewIdentityReceiptStore(pool)
	contractService := services.NewContractService(contractRepo, receiptStore)

	server := &Server{
		db:              pool,
		router:          chi.NewRouter(),
		contractService: contractService,
	}

	tests := []struct {
		name    string
		payload CreateContractRequest
	}{
		{
			name: "missing subject_domain_id",
			payload: CreateContractRequest{
				ContractType:  "tos",
				DSCIChannel:   "web",
				DSCIReference: "https://example.com/tos",
				DSCIVersion:   "1.0",
				EffectiveAt:   time.Now().UTC().Format(time.RFC3339),
			},
		},
		{
			name: "missing contract_type",
			payload: CreateContractRequest{
				SubjectDomainID: subjectID.String(),
				DSCIChannel:     "web",
				DSCIReference:   "https://example.com/tos",
				DSCIVersion:     "1.0",
				EffectiveAt:     time.Now().UTC().Format(time.RFC3339),
			},
		},
		{
			name: "missing dsci_channel",
			payload: CreateContractRequest{
				SubjectDomainID: subjectID.String(),
				ContractType:    "tos",
				DSCIReference:   "https://example.com/tos",
				DSCIVersion:     "1.0",
				EffectiveAt:     time.Now().UTC().Format(time.RFC3339),
			},
		},
		{
			name: "missing dsci_reference",
			payload: CreateContractRequest{
				SubjectDomainID: subjectID.String(),
				ContractType:    "tos",
				DSCIChannel:     "web",
				DSCIVersion:     "1.0",
				EffectiveAt:     time.Now().UTC().Format(time.RFC3339),
			},
		},
		{
			name: "missing dsci_version",
			payload: CreateContractRequest{
				SubjectDomainID: subjectID.String(),
				ContractType:    "tos",
				DSCIChannel:     "web",
				DSCIReference:   "https://example.com/tos",
				EffectiveAt:     time.Now().UTC().Format(time.RFC3339),
			},
		},
		{
			name: "missing effective_at",
			payload: CreateContractRequest{
				SubjectDomainID: subjectID.String(),
				ContractType:    "tos",
				DSCIChannel:     "web",
				DSCIReference:   "https://example.com/tos",
				DSCIVersion:     "1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/domain/"+issuerID.String()+"/contracts", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", issuerID.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			server.HandleCreateContract(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("%s: Expected status 400, got %d", tt.name, w.Code)
			}
		})
	}
}

func TestHandleGetDomainContracts(t *testing.T) {
	pool := setupContractAPITestDB(t)
	defer pool.Close()

	issuerID, subjectID := setupContractAPITestDomains(t, pool)
	defer cleanupContractAPIData(t, pool, issuerID, subjectID)

	contractRepo := repo.NewContractRepository(pool)
	receiptStore := identity.NewIdentityReceiptStore(pool)
	contractService := services.NewContractService(contractRepo, receiptStore)

	server := &Server{
		db:              pool,
		router:          chi.NewRouter(),
		contractService: contractService,
	}

	// Create test contracts
	ctx := context.Background()
	input1 := contracts.CreateContractInput{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "membership",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/membership",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
	}
	_, err := contractService.CreateContract(ctx, input1)
	if err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}

	input2 := contracts.CreateContractInput{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "tos",
		DSCIChannel:     "api",
		DSCIReference:   "https://example.com/tos",
		DSCIVersion:     "2.0.0",
		EffectiveAt:     time.Now().UTC().Add(-2 * time.Hour),
	}
	created2, err := contractService.CreateContract(ctx, input2)
	if err != nil {
		t.Fatalf("Failed to create second test contract: %v", err)
	}

	// Revoke second contract
	_, err = contractService.RevokeContract(ctx, created2.ID.String(), "test revocation", issuerID.String())
	if err != nil {
		t.Fatalf("Failed to revoke contract: %v", err)
	}

	// Test: List all contracts (no filter)
	t.Run("list all contracts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/domain/"+issuerID.String()+"/contracts", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", issuerID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		server.HandleGetDomainContracts(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response ContractListResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Count < 2 {
			t.Errorf("Expected at least 2 contracts, got %d", response.Count)
		}
	})

	// Test: Filter by status=active
	t.Run("filter by status active", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/domain/"+issuerID.String()+"/contracts?status=active", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", issuerID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		server.HandleGetDomainContracts(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response ContractListResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		for _, c := range response.Contracts {
			if c.Status != "active" {
				t.Errorf("Expected only active contracts, found %s", c.Status)
			}
		}
	})

	// Test: Filter by status=revoked
	t.Run("filter by status revoked", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/domain/"+issuerID.String()+"/contracts?status=revoked", nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", issuerID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		server.HandleGetDomainContracts(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response ContractListResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		for _, c := range response.Contracts {
			if c.Status != "revoked" {
				t.Errorf("Expected only revoked contracts, found %s", c.Status)
			}
		}
	})
}

func TestHandleGetContract_Success(t *testing.T) {
	pool := setupContractAPITestDB(t)
	defer pool.Close()

	issuerID, subjectID := setupContractAPITestDomains(t, pool)
	defer cleanupContractAPIData(t, pool, issuerID, subjectID)

	contractRepo := repo.NewContractRepository(pool)
	receiptStore := identity.NewIdentityReceiptStore(pool)
	contractService := services.NewContractService(contractRepo, receiptStore)

	server := &Server{
		db:              pool,
		router:          chi.NewRouter(),
		contractService: contractService,
	}

	// Create test contract
	ctx := context.Background()
	input := contracts.CreateContractInput{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "subscription",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/subscription",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
	}
	created, err := contractService.CreateContract(ctx, input)
	if err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}

	// Get contract
	req := httptest.NewRequest(http.MethodGet, "/api/domain/"+issuerID.String()+"/contracts/"+created.ID.String(), nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", issuerID.String())
	rctx.URLParams.Add("contractId", created.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	server.HandleGetContract(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response ContractDTO
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.ID != created.ID {
		t.Errorf("Expected contract ID %s, got %s", created.ID, response.ID)
	}
	if response.ContractType != "subscription" {
		t.Errorf("Expected contract_type 'subscription', got %s", response.ContractType)
	}
}

func TestHandleGetContract_NotFound(t *testing.T) {
	pool := setupContractAPITestDB(t)
	defer pool.Close()

	issuerID, subjectID := setupContractAPITestDomains(t, pool)
	defer cleanupContractAPIData(t, pool, issuerID, subjectID)

	contractRepo := repo.NewContractRepository(pool)
	receiptStore := identity.NewIdentityReceiptStore(pool)
	contractService := services.NewContractService(contractRepo, receiptStore)

	server := &Server{
		db:              pool,
		router:          chi.NewRouter(),
		contractService: contractService,
	}

	nonExistentID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/api/domain/"+issuerID.String()+"/contracts/"+nonExistentID, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", issuerID.String())
	rctx.URLParams.Add("contractId", nonExistentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	server.HandleGetContract(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleRevokeContract(t *testing.T) {
	pool := setupContractAPITestDB(t)
	defer pool.Close()

	issuerID, subjectID := setupContractAPITestDomains(t, pool)
	defer cleanupContractAPIData(t, pool, issuerID, subjectID)

	contractRepo := repo.NewContractRepository(pool)
	receiptStore := identity.NewIdentityReceiptStore(pool)
	contractService := services.NewContractService(contractRepo, receiptStore)

	server := &Server{
		db:              pool,
		router:          chi.NewRouter(),
		contractService: contractService,
	}

	// Create test contract
	ctx := context.Background()
	input := contracts.CreateContractInput{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "membership",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/membership",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
	}
	created, err := contractService.CreateContract(ctx, input)
	if err != nil {
		t.Fatalf("Failed to create test contract: %v", err)
	}

	// Revoke contract
	payload := RevokeContractRequest{
		Reason:  "User cancellation",
		ActorID: issuerID.String(),
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/domain/"+issuerID.String()+"/contracts/"+created.ID.String()+"/revoke", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", issuerID.String())
	rctx.URLParams.Add("contractId", created.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	server.HandleRevokeContract(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response ContractDTO
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "revoked" {
		t.Errorf("Expected status 'revoked', got %s", response.Status)
	}
	if response.RevokedAt == nil {
		t.Error("Expected revoked_at to be set")
	}
}
