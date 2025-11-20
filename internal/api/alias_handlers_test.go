package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dis-core/internal/domain"
	"dis-core/internal/identity"
	"dis-core/internal/repo"
	"dis-core/internal/testdb"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GOV-12: Alias Canon & DSCI Integration - API Handler Tests

// setupTestDomainsForAPI creates test domains
func setupTestDomainsForAPI(t *testing.T, pool *pgxpool.Pool) (ownerID, targetID uuid.UUID) {
	ctx := context.Background()

	ownerID = uuid.New()
	targetID = uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO domains (id, name, domain_type, authority, created_at)
		VALUES
			($1, 'api-test-owner', 'corporeal', 'test', NOW()),
			($2, 'api-test-target', 'organizational', 'test', NOW())
		ON CONFLICT (id) DO NOTHING
	`, ownerID, targetID)

	if err != nil {
		t.Fatalf("Failed to create test domains: %v", err)
	}

	return ownerID, targetID
}

// cleanupAliasesForAPI removes test aliases
func cleanupAliasesForAPI(t *testing.T, pool *pgxpool.Pool, aliasIDs ...uuid.UUID) {
	ctx := context.Background()
	for _, id := range aliasIDs {
		_, _ = pool.Exec(ctx, "DELETE FROM aliases WHERE id = $1", id)
	}
}

func TestHandleGetDomainAliases_Success(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	// Create test server
	server := &Server{
		db:                      pool,
		router:                  chi.NewRouter(),
		identityReceiptRecorder: identity.NewIdentityReceiptStore(pool),
	}

	ownerID, targetID := setupTestDomainsForAPI(t, pool)

	// Create test aliases
	aliasRepo := repo.NewAliasRepository(pool)
	alias1 := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: targetID,
		AliasName:      "test-alias-1",
		AliasType:      domain.AliasTypeRelationship,
		Metadata:       make(domain.AliasMetadata),
	}
	alias2 := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: ownerID,
		AliasName:      "test-mask-1",
		AliasType:      domain.AliasTypeMask,
		Metadata:       make(domain.AliasMetadata),
	}

	ctx := context.Background()
	if err := aliasRepo.CreateAlias(ctx, alias1); err != nil {
		t.Fatalf("Failed to create test alias: %v", err)
	}
	defer cleanupAliasesForAPI(t, pool, alias1.ID)

	if err := aliasRepo.CreateAlias(ctx, alias2); err != nil {
		t.Fatalf("Failed to create test alias: %v", err)
	}
	defer cleanupAliasesForAPI(t, pool, alias2.ID)

	// Make request
	req := httptest.NewRequest("GET", "/api/domain/"+ownerID.String()+"/aliases", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", ownerID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	server.HandleGetDomainAliases(rr, req)

	// Check response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response AliasListResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Count < 2 {
		t.Errorf("Expected at least 2 aliases, got %d", response.Count)
	}
}

func TestHandleGetDomainAliases_TypeFilter(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	server := &Server{
		db:                      pool,
		router:                  chi.NewRouter(),
		identityReceiptRecorder: identity.NewIdentityReceiptStore(pool),
	}

	ownerID, targetID := setupTestDomainsForAPI(t, pool)

	// Create test aliases of different types
	aliasRepo := repo.NewAliasRepository(pool)
	relAlias := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: targetID,
		AliasName:      "rel-filter-test",
		AliasType:      domain.AliasTypeRelationship,
		Metadata:       make(domain.AliasMetadata),
	}
	maskAlias := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: ownerID,
		AliasName:      "mask-filter-test",
		AliasType:      domain.AliasTypeMask,
		Metadata:       make(domain.AliasMetadata),
	}

	ctx := context.Background()
	if err := aliasRepo.CreateAlias(ctx, relAlias); err != nil {
		t.Fatalf("Failed to create test alias: %v", err)
	}
	defer cleanupAliasesForAPI(t, pool, relAlias.ID)

	if err := aliasRepo.CreateAlias(ctx, maskAlias); err != nil {
		t.Fatalf("Failed to create test alias: %v", err)
	}
	defer cleanupAliasesForAPI(t, pool, maskAlias.ID)

	// Request with MASK type filter
	req := httptest.NewRequest("GET", "/api/domain/"+ownerID.String()+"/aliases?type=MASK", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", ownerID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	server.HandleGetDomainAliases(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	var response AliasListResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// All returned aliases should be MASK type
	for _, alias := range response.Aliases {
		if alias.AliasType != "MASK" {
			t.Errorf("Expected only MASK aliases, got %s", alias.AliasType)
		}
	}
}

func TestHandleGetDomainAliases_InvalidDomain(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	server := &Server{
		db:                      pool,
		router:                  chi.NewRouter(),
		identityReceiptRecorder: identity.NewIdentityReceiptStore(pool),
	}

	// Request with invalid UUID
	req := httptest.NewRequest("GET", "/api/domain/invalid-uuid/aliases", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	server.HandleGetDomainAliases(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleCreateRelationshipAlias_Success(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	server := &Server{
		db:                      pool,
		router:                  chi.NewRouter(),
		identityReceiptRecorder: identity.NewIdentityReceiptStore(pool),
	}

	ownerID, targetID := setupTestDomainsForAPI(t, pool)

	// Create request
	reqBody := CreateRelationshipAliasRequest{
		TargetDomainID: targetID.String(),
		AliasName:      "test-handle",
		Metadata: map[string]interface{}{
			"display_name": "Test Handle",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/domain/"+ownerID.String()+"/alias/relationship", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", ownerID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	server.HandleCreateRelationshipAlias(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var response AliasDTO
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	defer cleanupAliasesForAPI(t, pool, response.ID)

	if response.AliasType != "RELATIONSHIP" {
		t.Errorf("Expected AliasType = RELATIONSHIP, got %s", response.AliasType)
	}
	if response.AliasName != "test-handle" {
		t.Errorf("Expected AliasName = 'test-handle', got '%s'", response.AliasName)
	}
}

func TestHandleCreateRelationshipAlias_MissingName(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	server := &Server{
		db:                      pool,
		router:                  chi.NewRouter(),
		identityReceiptRecorder: identity.NewIdentityReceiptStore(pool),
	}

	ownerID, targetID := setupTestDomainsForAPI(t, pool)

	// Create request without alias_name
	reqBody := CreateRelationshipAliasRequest{
		TargetDomainID: targetID.String(),
		// AliasName missing
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/domain/"+ownerID.String()+"/alias/relationship", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", ownerID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	server.HandleCreateRelationshipAlias(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleCreateRelationshipAlias_Conflict(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	server := &Server{
		db:                      pool,
		router:                  chi.NewRouter(),
		identityReceiptRecorder: identity.NewIdentityReceiptStore(pool),
	}

	ownerID, targetID := setupTestDomainsForAPI(t, pool)

	// Create first alias
	aliasRepo := repo.NewAliasRepository(pool)
	firstAlias := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: targetID,
		AliasName:      "conflict-test",
		AliasType:      domain.AliasTypeRelationship,
		Metadata:       make(domain.AliasMetadata),
	}

	ctx := context.Background()
	if err := aliasRepo.CreateAlias(ctx, firstAlias); err != nil {
		t.Fatalf("Failed to create first alias: %v", err)
	}
	defer cleanupAliasesForAPI(t, pool, firstAlias.ID)

	// Try to create duplicate
	reqBody := CreateRelationshipAliasRequest{
		TargetDomainID: targetID.String(),
		AliasName:      "conflict-test",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/domain/"+ownerID.String()+"/alias/relationship", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", ownerID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	server.HandleCreateRelationshipAlias(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", rr.Code)
	}
}

func TestHandleCreateMaskAlias_Success(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	server := &Server{
		db:                      pool,
		router:                  chi.NewRouter(),
		identityReceiptRecorder: identity.NewIdentityReceiptStore(pool),
	}

	ownerID, _ := setupTestDomainsForAPI(t, pool)

	// Create request with custom name
	ttlDays := 30
	reqBody := CreateMaskAliasRequest{
		RequestedName: "custom-mask",
		TTLDays:       &ttlDays,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/domain/"+ownerID.String()+"/alias/mask", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", ownerID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	server.HandleCreateMaskAlias(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var response AliasDTO
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	defer cleanupAliasesForAPI(t, pool, response.ID)

	if response.AliasType != "MASK" {
		t.Errorf("Expected AliasType = MASK, got %s", response.AliasType)
	}
	if response.AliasName != "custom-mask" {
		t.Errorf("Expected AliasName = 'custom-mask', got '%s'", response.AliasName)
	}
	if response.OwnerDomainID != ownerID {
		t.Error("Expected owner to match ownerID")
	}
	if response.TargetDomainID != ownerID {
		t.Error("Expected target to match ownerID (self-targeting)")
	}
}

func TestHandleCreateMaskAlias_AutoGenerateName(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	server := &Server{
		db:                      pool,
		router:                  chi.NewRouter(),
		identityReceiptRecorder: identity.NewIdentityReceiptStore(pool),
	}

	ownerID, _ := setupTestDomainsForAPI(t, pool)

	// Create request without requested_name
	reqBody := CreateMaskAliasRequest{
		// RequestedName empty - should auto-generate
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/domain/"+ownerID.String()+"/alias/mask", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", ownerID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	server.HandleCreateMaskAlias(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var response AliasDTO
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	defer cleanupAliasesForAPI(t, pool, response.ID)

	// Verify name was generated
	if len(response.AliasName) < 5 || response.AliasName[:5] != "mask-" {
		t.Errorf("Expected auto-generated name starting with 'mask-', got '%s'", response.AliasName)
	}
}

func TestHandleCreateMaskAlias_InvalidJSON(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	server := &Server{
		db:                      pool,
		router:                  chi.NewRouter(),
		identityReceiptRecorder: identity.NewIdentityReceiptStore(pool),
	}

	ownerID, _ := setupTestDomainsForAPI(t, pool)

	// Send invalid JSON
	req := httptest.NewRequest("POST", "/api/domain/"+ownerID.String()+"/alias/mask", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", ownerID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	server.HandleCreateMaskAlias(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}
