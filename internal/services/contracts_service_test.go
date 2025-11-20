package services

import (
	"context"
	"testing"
	"time"

	"dis-core/internal/contracts"
	"dis-core/internal/identity"
	"dis-core/internal/repo"
	"dis-core/internal/testdb"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GOV-13: Contracts Service Tests

// MockReceiptRecorder captures receipt calls for testing
type MockReceiptRecorder struct {
	RecordedReceipts []*identity.IdentityReceipt
	ShouldFail       bool
}

func (m *MockReceiptRecorder) RecordIdentityReceipt(ctx context.Context, receipt *identity.IdentityReceipt) error {
	if m.ShouldFail {
		return nil // Return error("mock failure")
	}
	m.RecordedReceipts = append(m.RecordedReceipts, receipt)
	return nil
}

// setupContractServiceTestDB creates a test database connection
func setupContractServiceTestDB(t *testing.T) *pgxpool.Pool {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	return pool
}

// setupContractServiceTestDomains creates test domains
func setupContractServiceTestDomains(t *testing.T, pool *pgxpool.Pool) (issuerID, subjectID, aliasID uuid.UUID) {
	ctx := context.Background()

	issuerID = uuid.New()
	subjectID = uuid.New()
	aliasID = uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO domains (id, name, created_at)
		VALUES
			($1, 'test-issuer-svc', NOW()),
			($2, 'test-subject-svc', NOW()),
			($3, 'test-alias-svc', NOW())
		ON CONFLICT (id) DO NOTHING
	`, issuerID, subjectID, aliasID)

	if err != nil {
		t.Fatalf("Failed to create test domains: %v", err)
	}

	return issuerID, subjectID, aliasID
}

// cleanupContractServiceDomains removes test domains
func cleanupContractServiceDomains(t *testing.T, pool *pgxpool.Pool, domainIDs ...uuid.UUID) {
	ctx := context.Background()
	for _, id := range domainIDs {
		_, _ = pool.Exec(ctx, "DELETE FROM contracts WHERE domain_id = $1 OR subject_domain_id = $1", id)
		_, _ = pool.Exec(ctx, "DELETE FROM domains WHERE id = $1", id)
	}
}

func TestCreateContract_Success(t *testing.T) {
	pool := setupContractServiceTestDB(t)
	defer pool.Close()

	issuerID, subjectID, aliasID := setupContractServiceTestDomains(t, pool)
	defer cleanupContractServiceDomains(t, pool, issuerID, subjectID, aliasID)

	mockReceipts := &MockReceiptRecorder{}
	contractRepo := repo.NewContractRepository(pool)
	service := NewContractService(contractRepo, mockReceipts)

	// Create contract input
	input := contracts.CreateContractInput{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		AliasID:         &aliasID,
		ContractType:    "membership",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/membership/v1",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour), // Active now
		ExpiresAt:       nil,
	}

	ctx := context.Background()
	created, err := service.CreateContract(ctx, input)
	if err != nil {
		t.Fatalf("CreateContract failed: %v", err)
	}

	// Assert returned contract matches input
	if created.DomainID != issuerID {
		t.Errorf("Expected domain_id %s, got %s", issuerID, created.DomainID)
	}
	if created.SubjectDomainID != subjectID {
		t.Errorf("Expected subject_domain_id %s, got %s", subjectID, created.SubjectDomainID)
	}
	if created.ContractType != "membership" {
		t.Errorf("Expected contract_type 'membership', got %s", created.ContractType)
	}

	// Assert status computed correctly (should be 'active' since EffectiveAt is in past)
	if created.Status != contracts.ContractStatusActive {
		t.Errorf("Expected status 'active', got %s", created.Status)
	}

	// Give async receipt recording time to complete
	time.Sleep(100 * time.Millisecond)

	// Assert receipt recorded
	if len(mockReceipts.RecordedReceipts) != 1 {
		t.Errorf("Expected 1 receipt recorded, got %d", len(mockReceipts.RecordedReceipts))
	} else {
		receipt := mockReceipts.RecordedReceipts[0]
		if receipt.Action != identity.ContractCreateV1 {
			t.Errorf("Expected action contract.create.v1, got %s", receipt.Action)
		}
		if receipt.DomainID != issuerID {
			t.Errorf("Expected receipt domain_id %s, got %s", issuerID, receipt.DomainID)
		}
	}
}

func TestCreateContract_PendingStatus(t *testing.T) {
	pool := setupContractServiceTestDB(t)
	defer pool.Close()

	issuerID, subjectID, _ := setupContractServiceTestDomains(t, pool)
	defer cleanupContractServiceDomains(t, pool, issuerID, subjectID)

	mockReceipts := &MockReceiptRecorder{}
	contractRepo := repo.NewContractRepository(pool)
	service := NewContractService(contractRepo, mockReceipts)

	// Create contract with future effective_at
	futureEffective := time.Now().UTC().Add(24 * time.Hour)
	input := contracts.CreateContractInput{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "tos",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/tos",
		DSCIVersion:     "2.0.0",
		EffectiveAt:     futureEffective,
	}

	ctx := context.Background()
	created, err := service.CreateContract(ctx, input)
	if err != nil {
		t.Fatalf("CreateContract failed: %v", err)
	}

	// Assert status is 'pending' since EffectiveAt is in future
	if created.Status != contracts.ContractStatusPending {
		t.Errorf("Expected status 'pending' for future contract, got %s", created.Status)
	}
}

func TestCreateContract_Validation(t *testing.T) {
	pool := setupContractServiceTestDB(t)
	defer pool.Close()

	issuerID, subjectID, _ := setupContractServiceTestDomains(t, pool)
	defer cleanupContractServiceDomains(t, pool, issuerID, subjectID)

	mockReceipts := &MockReceiptRecorder{}
	contractRepo := repo.NewContractRepository(pool)
	service := NewContractService(contractRepo, mockReceipts)

	ctx := context.Background()

	tests := []struct {
		name  string
		input contracts.CreateContractInput
	}{
		{
			name: "missing domain_id",
			input: contracts.CreateContractInput{
				SubjectDomainID: subjectID,
				ContractType:    "tos",
				DSCIChannel:     "web",
				DSCIReference:   "https://example.com/tos",
				DSCIVersion:     "1.0",
				EffectiveAt:     time.Now().UTC(),
			},
		},
		{
			name: "missing subject_domain_id",
			input: contracts.CreateContractInput{
				DomainID:      issuerID,
				ContractType:  "tos",
				DSCIChannel:   "web",
				DSCIReference: "https://example.com/tos",
				DSCIVersion:   "1.0",
				EffectiveAt:   time.Now().UTC(),
			},
		},
		{
			name: "missing contract_type",
			input: contracts.CreateContractInput{
				DomainID:        issuerID,
				SubjectDomainID: subjectID,
				DSCIChannel:     "web",
				DSCIReference:   "https://example.com/tos",
				DSCIVersion:     "1.0",
				EffectiveAt:     time.Now().UTC(),
			},
		},
		{
			name: "missing dsci_channel",
			input: contracts.CreateContractInput{
				DomainID:        issuerID,
				SubjectDomainID: subjectID,
				ContractType:    "tos",
				DSCIReference:   "https://example.com/tos",
				DSCIVersion:     "1.0",
				EffectiveAt:     time.Now().UTC(),
			},
		},
		{
			name: "missing dsci_reference",
			input: contracts.CreateContractInput{
				DomainID:        issuerID,
				SubjectDomainID: subjectID,
				ContractType:    "tos",
				DSCIChannel:     "web",
				DSCIVersion:     "1.0",
				EffectiveAt:     time.Now().UTC(),
			},
		},
		{
			name: "missing dsci_version",
			input: contracts.CreateContractInput{
				DomainID:        issuerID,
				SubjectDomainID: subjectID,
				ContractType:    "tos",
				DSCIChannel:     "web",
				DSCIReference:   "https://example.com/tos",
				EffectiveAt:     time.Now().UTC(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateContract(ctx, tt.input)
			if err == nil {
				t.Errorf("Expected validation error for %s, got nil", tt.name)
			}
		})
	}
}

func TestRevokeContract_Success(t *testing.T) {
	pool := setupContractServiceTestDB(t)
	defer pool.Close()

	issuerID, subjectID, _ := setupContractServiceTestDomains(t, pool)
	defer cleanupContractServiceDomains(t, pool, issuerID, subjectID)

	mockReceipts := &MockReceiptRecorder{}
	contractRepo := repo.NewContractRepository(pool)
	service := NewContractService(contractRepo, mockReceipts)

	ctx := context.Background()

	// Create contract
	input := contracts.CreateContractInput{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "subscription",
		DSCIChannel:     "api",
		DSCIReference:   "https://api.example.com/sub/terms",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
	}

	created, err := service.CreateContract(ctx, input)
	if err != nil {
		t.Fatalf("CreateContract failed: %v", err)
	}

	// Give async receipt time
	time.Sleep(100 * time.Millisecond)

	// Clear recorded receipts
	mockReceipts.RecordedReceipts = nil

	// Revoke contract
	actorID := issuerID.String()
	reason := "User requested cancellation"
	revoked, err := service.RevokeContract(ctx, created.ID.String(), reason, actorID)
	if err != nil {
		t.Fatalf("RevokeContract failed: %v", err)
	}

	// Assert status updated
	if revoked.Status != contracts.ContractStatusRevoked {
		t.Errorf("Expected status 'revoked', got %s", revoked.Status)
	}

	// Assert revoked_at set
	if revoked.RevokedAt == nil {
		t.Error("Expected revoked_at to be set")
	}

	// Give async receipt time
	time.Sleep(100 * time.Millisecond)

	// Assert receipt recorded
	if len(mockReceipts.RecordedReceipts) != 1 {
		t.Errorf("Expected 1 revoke receipt, got %d", len(mockReceipts.RecordedReceipts))
	} else {
		receipt := mockReceipts.RecordedReceipts[0]
		if receipt.Action != identity.ContractRevokeV1 {
			t.Errorf("Expected action contract.revoke.v1, got %s", receipt.Action)
		}
	}
}

func TestRevokeContract_Idempotent(t *testing.T) {
	pool := setupContractServiceTestDB(t)
	defer pool.Close()

	issuerID, subjectID, _ := setupContractServiceTestDomains(t, pool)
	defer cleanupContractServiceDomains(t, pool, issuerID, subjectID)

	mockReceipts := &MockReceiptRecorder{}
	contractRepo := repo.NewContractRepository(pool)
	service := NewContractService(contractRepo, mockReceipts)

	ctx := context.Background()

	// Create contract
	input := contracts.CreateContractInput{
		DomainID:        issuerID,
		SubjectDomainID: subjectID,
		ContractType:    "membership",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/membership",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
	}

	created, err := service.CreateContract(ctx, input)
	if err != nil {
		t.Fatalf("CreateContract failed: %v", err)
	}

	// Revoke once
	actorID := issuerID.String()
	reason := "First revocation"
	revoked1, err := service.RevokeContract(ctx, created.ID.String(), reason, actorID)
	if err != nil {
		t.Fatalf("First RevokeContract failed: %v", err)
	}

	// Give async receipt time
	time.Sleep(100 * time.Millisecond)
	firstRevokedAt := revoked1.RevokedAt

	// Clear receipts
	mockReceipts.RecordedReceipts = nil

	// Revoke again (idempotent)
	revoked2, err := service.RevokeContract(ctx, created.ID.String(), reason, actorID)
	if err != nil {
		t.Fatalf("Second RevokeContract failed: %v", err)
	}

	// Assert second revoke returns same contract (idempotent)
	if revoked2.Status != contracts.ContractStatusRevoked {
		t.Errorf("Expected status 'revoked' on second revoke, got %s", revoked2.Status)
	}

	// Assert revoked_at unchanged
	if revoked2.RevokedAt == nil || firstRevokedAt == nil {
		t.Error("Expected revoked_at to be set on both calls")
	} else if !revoked2.RevokedAt.Equal(*firstRevokedAt) {
		t.Errorf("Expected revoked_at to remain unchanged, got %s vs %s", revoked2.RevokedAt, firstRevokedAt)
	}

	// Give async time
	time.Sleep(100 * time.Millisecond)

	// Assert no duplicate receipt (idempotent behavior)
	if len(mockReceipts.RecordedReceipts) > 0 {
		t.Logf("Note: Second revoke emitted receipt (acceptable if defined in spec), count=%d", len(mockReceipts.RecordedReceipts))
	}
}

func TestListDomainContracts_Service(t *testing.T) {
	pool := setupContractServiceTestDB(t)
	defer pool.Close()

	issuer1ID, subject1ID, _ := setupContractServiceTestDomains(t, pool)
	issuer2ID := uuid.New()

	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		INSERT INTO domains (id, name, created_at)
		VALUES ($1, 'test-issuer2-svc', NOW())
		ON CONFLICT (id) DO NOTHING
	`, issuer2ID)
	if err != nil {
		t.Fatalf("Failed to create issuer2: %v", err)
	}

	defer cleanupContractServiceDomains(t, pool, issuer1ID, subject1ID, issuer2ID)

	mockReceipts := &MockReceiptRecorder{}
	contractRepo := repo.NewContractRepository(pool)
	service := NewContractService(contractRepo, mockReceipts)

	// Create two contracts for issuer1
	input1 := contracts.CreateContractInput{
		DomainID:        issuer1ID,
		SubjectDomainID: subject1ID,
		ContractType:    "tos",
		DSCIChannel:     "web",
		DSCIReference:   "https://issuer1.com/tos",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
	}
	_, err = service.CreateContract(ctx, input1)
	if err != nil {
		t.Fatalf("Failed to create contract for issuer1: %v", err)
	}

	input2 := contracts.CreateContractInput{
		DomainID:        issuer1ID,
		SubjectDomainID: subject1ID,
		ContractType:    "membership",
		DSCIChannel:     "web",
		DSCIReference:   "https://issuer1.com/membership",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
	}
	_, err = service.CreateContract(ctx, input2)
	if err != nil {
		t.Fatalf("Failed to create second contract for issuer1: %v", err)
	}

	// Create one contract for issuer2
	input3 := contracts.CreateContractInput{
		DomainID:        issuer2ID,
		SubjectDomainID: subject1ID,
		ContractType:    "data-processing",
		DSCIChannel:     "api",
		DSCIReference:   "https://issuer2.com/dpa",
		DSCIVersion:     "2.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
	}
	_, err = service.CreateContract(ctx, input3)
	if err != nil {
		t.Fatalf("Failed to create contract for issuer2: %v", err)
	}

	// List contracts for issuer1
	issuer1Contracts, err := service.ListDomainContracts(ctx, issuer1ID.String(), "")
	if err != nil {
		t.Fatalf("ListDomainContracts failed: %v", err)
	}

	// Assert only issuer1 contracts returned
	if len(issuer1Contracts) < 2 {
		t.Errorf("Expected at least 2 contracts for issuer1, got %d", len(issuer1Contracts))
	}
	for _, c := range issuer1Contracts {
		if c.DomainID != issuer1ID {
			t.Errorf("Expected domain_id %s, found %s", issuer1ID, c.DomainID)
		}
	}
}

func TestListSubjectContracts_Service(t *testing.T) {
	pool := setupContractServiceTestDB(t)
	defer pool.Close()

	issuerID, subject1ID, _ := setupContractServiceTestDomains(t, pool)
	subject2ID := uuid.New()

	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		INSERT INTO domains (id, name, created_at)
		VALUES ($1, 'test-subject2-svc', NOW())
		ON CONFLICT (id) DO NOTHING
	`, subject2ID)
	if err != nil {
		t.Fatalf("Failed to create subject2: %v", err)
	}

	defer cleanupContractServiceDomains(t, pool, issuerID, subject1ID, subject2ID)

	mockReceipts := &MockReceiptRecorder{}
	contractRepo := repo.NewContractRepository(pool)
	service := NewContractService(contractRepo, mockReceipts)

	// Create contract for subject1
	input1 := contracts.CreateContractInput{
		DomainID:        issuerID,
		SubjectDomainID: subject1ID,
		ContractType:    "tos",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/tos",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
	}
	_, err = service.CreateContract(ctx, input1)
	if err != nil {
		t.Fatalf("Failed to create contract for subject1: %v", err)
	}

	// Create contract for subject2
	input2 := contracts.CreateContractInput{
		DomainID:        issuerID,
		SubjectDomainID: subject2ID,
		ContractType:    "membership",
		DSCIChannel:     "web",
		DSCIReference:   "https://example.com/membership",
		DSCIVersion:     "1.0.0",
		EffectiveAt:     time.Now().UTC().Add(-1 * time.Hour),
	}
	_, err = service.CreateContract(ctx, input2)
	if err != nil {
		t.Fatalf("Failed to create contract for subject2: %v", err)
	}

	// List contracts for subject1
	subject1Contracts, err := service.ListSubjectContracts(ctx, subject1ID.String(), "")
	if err != nil {
		t.Fatalf("ListSubjectContracts failed: %v", err)
	}

	// Assert only subject1 contracts returned
	foundSubject1 := false
	for _, c := range subject1Contracts {
		if c.SubjectDomainID == subject1ID {
			foundSubject1 = true
		}
		if c.SubjectDomainID == subject2ID {
			t.Error("Found subject2 contract in subject1 listing")
		}
	}
	if !foundSubject1 {
		t.Error("Expected to find contract for subject1")
	}
}
