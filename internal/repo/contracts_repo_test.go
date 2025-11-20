package repo

import (
	"context"
	"testing"
	"time"

	"dis-core/internal/contracts"
	"dis-core/internal/testdb"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GOV-13: Contracts Table & DSCI Contract Wiring - Repository Tests

// setupContractTestDB creates a test database connection
func setupContractTestDB(t *testing.T) *pgxpool.Pool {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	return pool
}

// setupContractTestDomains creates test domains for contract testing
func setupContractTestDomains(t *testing.T, pool *pgxpool.Pool) (issuerID, subjectID, aliasID uuid.UUID) {
	ctx := context.Background()

	issuerID = uuid.New()
	subjectID = uuid.New()
	aliasID = uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO domains (id, name, created_at)
		VALUES
			($1, 'test-issuer-domain', NOW()),
			($2, 'test-subject-domain', NOW()),
			($3, 'test-alias-domain', NOW())
		ON CONFLICT (id) DO NOTHING
	`, issuerID, subjectID, aliasID)

	if err != nil {
		t.Fatalf("Failed to create test domains: %v", err)
	}

	return issuerID, subjectID, aliasID
}

// cleanupContracts removes test contracts
func cleanupContracts(t *testing.T, pool *pgxpool.Pool, contractIDs ...uuid.UUID) {
	ctx := context.Background()
	for _, id := range contractIDs {
		_, _ = pool.Exec(ctx, "DELETE FROM contracts WHERE id = $1", id)
	}
}

// cleanupContractDomains removes test domains
func cleanupContractDomains(t *testing.T, pool *pgxpool.Pool, domainIDs ...uuid.UUID) {
	ctx := context.Background()
	for _, id := range domainIDs {
		_, _ = pool.Exec(ctx, "DELETE FROM domains WHERE id = $1", id)
	}
}

func TestCreateContract_Success(t *testing.T) {
	pool := setupContractTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	issuerID, subjectID, aliasID := setupContractTestDomains(t, pool)
	defer cleanupContractDomains(t, pool, issuerID, subjectID, aliasID)

	repo := NewContractRepository(pool)

	// Create contract with full payload
	contract := &contracts.Contract{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		AliasID:         &aliasID,
		ContractType:    "membership",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/tos/v1",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour), // Effective now
		ExpiresAt:       nil,
		Status:          contracts.ContractStatusActive,
	}

	created, err := repo.CreateContract(ctx, contract)
	if err != nil {
		t.Fatalf("CreateContract failed: %v", err)
	}
	defer cleanupContracts(t, pool, created.ID)

	// Assert returned contract has generated fields
	if created.ID == uuid.Nil {
		t.Error("Expected non-nil ID")
	}
	if created.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if created.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
	if created.Status != contracts.ContractStatusActive {
		t.Errorf("Expected status 'active', got %s", created.Status)
	}

	// Verify contract can be retrieved
	retrieved, err := repo.GetContractByID(ctx, created.ID.String())
	if err != nil {
		t.Fatalf("GetContractByID failed: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
	}
	if retrieved.ContractType != "membership" {
		t.Errorf("Expected contract_type 'membership', got %s", retrieved.ContractType)
	}
	if retrieved.DSCIChannel != "web" {
		t.Errorf("Expected dsci_channel 'web', got %s", retrieved.DSCIChannel)
	}
}

func TestCreateContract_RequiredFields(t *testing.T) {
	pool := setupContractTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	issuerID, subjectID, _ := setupContractTestDomains(t, pool)
	defer cleanupContractDomains(t, pool, issuerID, subjectID)

	repo := NewContractRepository(pool)

	tests := []struct {
		name        string
		contract    *contracts.Contract
		expectError bool
	}{
		{
			name: "missing subject_domain_id",
			contract: &contracts.Contract{
				DomainID:      issuerID,
				ContractType:  "tos",
				DSCIChannel:   "web",
				DSCIReference: "https://example.com/tos",
				DSCIVersion:   "1.0",
				EffectiveAt:   time.Now().UTC(),
				Status:        contracts.ContractStatusActive,
			},
			expectError: true,
		},
		{
			name: "missing contract_type",
			contract: &contracts.Contract{
				DomainID:        issuerID,
				SubjectDomainID: subjectID,
				DSCIChannel:     "web",
				DSCIReference:   "https://example.com/tos",
				DSCIVersion:     "1.0",
				EffectiveAt:     time.Now().UTC(),
				Status:          contracts.ContractStatusActive,
			},
			expectError: true,
		},
		{
			name: "missing dsci_reference",
			contract: &contracts.Contract{
				DomainID:        issuerID,
				SubjectDomainID: subjectID,
				ContractType:    "tos",
				DSCIChannel:     "web",
				DSCIVersion:     "1.0",
				EffectiveAt:     time.Now().UTC(),
				Status:          contracts.ContractStatusActive,
			},
			expectError: true,
		},
		{
			name: "invalid status",
			contract: &contracts.Contract{
				DomainID:        issuerID,
				SubjectDomainID: subjectID,
				ContractType:    "tos",
				DSCIChannel:     "web",
				DSCIReference:   "https://example.com/tos",
				DSCIVersion:     "1.0",
				EffectiveAt:     time.Now().UTC(),
				Status:          "invalid_status",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repo.CreateContract(ctx, tt.contract)
			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Did not expect error for %s, got: %v", tt.name, err)
			}
		})
	}
}

func TestGetContractByID_NotFound(t *testing.T) {
	pool := setupContractTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	repo := NewContractRepository(pool)

	nonExistentID := uuid.New().String()
	_, err := repo.GetContractByID(ctx, nonExistentID)
	if err == nil {
		t.Error("Expected error for non-existent contract, got nil")
	}
}

func TestListContractsByDomain(t *testing.T) {
	pool := setupContractTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	issuerID, subjectID, _ := setupContractTestDomains(t, pool)
	defer cleanupContractDomains(t, pool, issuerID, subjectID)

	repo := NewContractRepository(pool)

	// Create multiple contracts for same issuer
	contract1 := &contracts.Contract{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "membership",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/membership",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
		Status:          contracts.ContractStatusActive,
	}

	revokedAt := time.Now().UTC()
	contract2 := &contracts.Contract{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "tos",
		DSCIChannel:     "api",
		DSCIReference:   "https://example.com/tos",
		DSCIVersion:     "2.0.0",
		EffectiveAt:     time.Now().UTC().Add(-2 * time.Hour),
		RevokedAt:       &revokedAt,
		Status:          contracts.ContractStatusRevoked,
	}

	created1, err := repo.CreateContract(ctx, contract1)
	if err != nil {
		t.Fatalf("Failed to create contract1: %v", err)
	}
	defer cleanupContracts(t, pool, created1.ID)

	created2, err := repo.CreateContract(ctx, contract2)
	if err != nil {
		t.Fatalf("Failed to create contract2: %v", err)
	}
	defer cleanupContracts(t, pool, created2.ID)

	// List all contracts for issuer
	allContracts, err := repo.ListContractsByDomain(ctx, issuerID.String(), "")
	if err != nil {
		t.Fatalf("ListContractsByDomain failed: %v", err)
	}

	if len(allContracts) < 2 {
		t.Errorf("Expected at least 2 contracts, got %d", len(allContracts))
	}

	// List only active contracts
	activeContracts, err := repo.ListContractsByDomain(ctx, issuerID.String(), "active")
	if err != nil {
		t.Fatalf("ListContractsByDomain (active filter) failed: %v", err)
	}

	foundActive := false
	for _, c := range activeContracts {
		if c.ID == created1.ID {
			foundActive = true
		}
		if c.Status != contracts.ContractStatusActive {
			t.Errorf("Expected only active contracts, found %s", c.Status)
		}
	}
	if !foundActive {
		t.Error("Expected to find active contract in filtered results")
	}

	// List only revoked contracts
	revokedContracts, err := repo.ListContractsByDomain(ctx, issuerID.String(), "revoked")
	if err != nil {
		t.Fatalf("ListContractsByDomain (revoked filter) failed: %v", err)
	}

	foundRevoked := false
	for _, c := range revokedContracts {
		if c.ID == created2.ID {
			foundRevoked = true
		}
		if c.Status != contracts.ContractStatusRevoked {
			t.Errorf("Expected only revoked contracts, found %s", c.Status)
		}
	}
	if !foundRevoked {
		t.Error("Expected to find revoked contract in filtered results")
	}
}

func TestListContractsBySubject(t *testing.T) {
	pool := setupContractTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	issuerID, subjectID, _ := setupContractTestDomains(t, pool)
	defer cleanupContractDomains(t, pool, issuerID, subjectID)

	repo := NewContractRepository(pool)

	// Create contract with subject
	contract := &contracts.Contract{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "data-processing",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/dpa",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
		Status:          contracts.ContractStatusActive,
	}

	created, err := repo.CreateContract(ctx, contract)
	if err != nil {
		t.Fatalf("Failed to create contract: %v", err)
	}
	defer cleanupContracts(t, pool, created.ID)

	// List contracts by subject
	subjectContracts, err := repo.ListContractsBySubject(ctx, subjectID.String(), "")
	if err != nil {
		t.Fatalf("ListContractsBySubject failed: %v", err)
	}

	found := false
	for _, c := range subjectContracts {
		if c.ID == created.ID {
			found = true
			if c.SubjectDomainID != subjectID {
				t.Errorf("Expected subject_domain_id %s, got %s", subjectID, c.SubjectDomainID)
			}
		}
	}
	if !found {
		t.Error("Expected to find contract in subject listing")
	}

	// Test with status filter
	activeSubjectContracts, err := repo.ListContractsBySubject(ctx, subjectID.String(), "active")
	if err != nil {
		t.Fatalf("ListContractsBySubject (active filter) failed: %v", err)
	}

	for _, c := range activeSubjectContracts {
		if c.Status != contracts.ContractStatusActive {
			t.Errorf("Expected only active contracts, found %s", c.Status)
		}
	}
}

func TestUpdateContractStatus(t *testing.T) {
	pool := setupContractTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	issuerID, subjectID, _ := setupContractTestDomains(t, pool)
	defer cleanupContractDomains(t, pool, issuerID, subjectID)

	repo := NewContractRepository(pool)

	// Create active contract
	contract := &contracts.Contract{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "subscription",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/subscription",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
		Status:          contracts.ContractStatusActive,
	}

	created, err := repo.CreateContract(ctx, contract)
	if err != nil {
		t.Fatalf("Failed to create contract: %v", err)
	}
	defer cleanupContracts(t, pool, created.ID)

	// Update to revoked status
	revokedAt := time.Now().UTC()
	updated, err := repo.UpdateContractStatus(ctx, created.ID.String(), contracts.ContractStatusRevoked, &revokedAt)
	if err != nil {
		t.Fatalf("UpdateContractStatus failed: %v", err)
	}

	// Assert status updated
	if updated.Status != contracts.ContractStatusRevoked {
		t.Errorf("Expected status 'revoked', got %s", updated.Status)
	}

	// Assert revoked_at set
	if updated.RevokedAt == nil {
		t.Error("Expected revoked_at to be set")
	} else {
		// Allow 1 second tolerance for timestamp comparison
		diff := updated.RevokedAt.Sub(revokedAt).Abs()
		if diff > time.Second {
			t.Errorf("Expected revoked_at near %s, got %s", revokedAt, *updated.RevokedAt)
		}
	}

	// Assert updated_at changed
	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Error("Expected updated_at to be after original created_at")
	}

	// Verify retrieval shows updated status
	retrieved, err := repo.GetContractByID(ctx, created.ID.String())
	if err != nil {
		t.Fatalf("GetContractByID failed: %v", err)
	}

	if retrieved.Status != contracts.ContractStatusRevoked {
		t.Errorf("Retrieved contract has wrong status: %s", retrieved.Status)
	}
}

func TestUpdateContractStatus_InvalidTransition(t *testing.T) {
	pool := setupContractTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	issuerID, subjectID, _ := setupContractTestDomains(t, pool)
	defer cleanupContractDomains(t, pool, issuerID, subjectID)

	repo := NewContractRepository(pool)

	// Create active contract
	contract := &contracts.Contract{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "tos",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/tos",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
		Status:          contracts.ContractStatusActive,
	}

	created, err := repo.CreateContract(ctx, contract)
	if err != nil {
		t.Fatalf("Failed to create contract: %v", err)
	}
	defer cleanupContracts(t, pool, created.ID)

	// Try to set revoked status without revoked_at
	_, err = repo.UpdateContractStatus(ctx, created.ID.String(), contracts.ContractStatusRevoked, nil)
	if err == nil {
		t.Error("Expected error when setting revoked status without revoked_at")
	}

	// Try to set active status with revoked_at
	revokedAt := time.Now().UTC()
	_, err = repo.UpdateContractStatus(ctx, created.ID.String(), contracts.ContractStatusActive, &revokedAt)
	if err == nil {
		t.Error("Expected error when setting active status with revoked_at")
	}
}
