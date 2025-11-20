package services

import (
	"context"
	"testing"

	"dis-core/internal/domain"
	"dis-core/internal/repo"
	"dis-core/internal/testdb"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GOV-12: Alias Canon & DSCI Integration - Service Layer Tests

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *pgxpool.Pool {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	return pool
}

// setupTestDomains creates test domains
func setupTestDomains(t *testing.T, pool *pgxpool.Pool) (corporealID, otherID uuid.UUID) {
	ctx := context.Background()

	corporealID = uuid.New()
	otherID = uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO domains (id, name, created_at)
		VALUES
			($1, 'test-corporeal', NOW()),
			($2, 'test-other', NOW())
		ON CONFLICT (id) DO NOTHING
	`, corporealID, otherID)

	if err != nil {
		t.Fatalf("Failed to create test domains: %v", err)
	}

	return corporealID, otherID
}

// cleanupAliases removes test aliases
func cleanupAliases(t *testing.T, pool *pgxpool.Pool, aliasIDs ...uuid.UUID) {
	ctx := context.Background()
	for _, id := range aliasIDs {
		_, _ = pool.Exec(ctx, "DELETE FROM aliases WHERE id = $1", id)
	}
}

func TestEnsureCorporealRootAlias_Idempotent(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	aliasRepo := repo.NewAliasRepository(pool)
	service := NewAliasService(aliasRepo)

	corporealID, _ := setupTestDomains(t, pool)

	// First call: should create
	alias1, err := service.EnsureCorporealRootAlias(ctx, corporealID)
	if err != nil {
		t.Fatalf("EnsureCorporealRootAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias1.ID)

	// Verify it's an AUTO alias
	if alias1.AliasType != domain.AliasTypeAuto {
		t.Errorf("Expected AliasType = AUTO, got %s", alias1.AliasType)
	}
	if !alias1.IsCorporealAuto {
		t.Error("Expected IsCorporealAuto = true, got false")
	}
	if alias1.OwnerDomainID != corporealID {
		t.Error("Expected owner to be corporeal domain")
	}
	if alias1.TargetDomainID != corporealID {
		t.Error("Expected target to be corporeal domain (self-targeting)")
	}

	// Second call: should return existing
	alias2, err := service.EnsureCorporealRootAlias(ctx, corporealID)
	if err != nil {
		t.Fatalf("EnsureCorporealRootAlias second call failed: %v", err)
	}

	// Should return the same alias
	if alias2.ID != alias1.ID {
		t.Errorf("Expected same alias ID, got different: %s != %s", alias2.ID, alias1.ID)
	}
}

func TestEnsureStructuralAlias_Idempotent(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	aliasRepo := repo.NewAliasRepository(pool)
	service := NewAliasService(aliasRepo)

	ownerID, targetID := setupTestDomains(t, pool)

	// First call: should create
	alias1, err := service.EnsureStructuralAlias(ctx, ownerID, targetID)
	if err != nil {
		t.Fatalf("EnsureStructuralAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias1.ID)

	// Verify it's an AUTO alias
	if alias1.AliasType != domain.AliasTypeAuto {
		t.Errorf("Expected AliasType = AUTO, got %s", alias1.AliasType)
	}
	if alias1.IsCorporealAuto {
		t.Error("Expected IsCorporealAuto = false for structural alias")
	}
	if alias1.OwnerDomainID != ownerID {
		t.Error("Expected owner to match ownerID")
	}
	if alias1.TargetDomainID != targetID {
		t.Error("Expected target to match targetID")
	}

	// Second call: should return existing
	alias2, err := service.EnsureStructuralAlias(ctx, ownerID, targetID)
	if err != nil {
		t.Fatalf("EnsureStructuralAlias second call failed: %v", err)
	}

	// Should return the same alias
	if alias2.ID != alias1.ID {
		t.Errorf("Expected same alias ID, got different: %s != %s", alias2.ID, alias1.ID)
	}
}

func TestCreateRelationshipAlias_Success(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	aliasRepo := repo.NewAliasRepository(pool)
	service := NewAliasService(aliasRepo)

	ownerID, targetID := setupTestDomains(t, pool)

	metadata := domain.AliasMetadata{
		"display_name": "My Handle",
	}

	alias, err := service.CreateRelationshipAlias(ctx, ownerID, targetID, "myhandle", nil, metadata)
	if err != nil {
		t.Fatalf("CreateRelationshipAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias.ID)

	// Verify it's a RELATIONSHIP alias
	if alias.AliasType != domain.AliasTypeRelationship {
		t.Errorf("Expected AliasType = RELATIONSHIP, got %s", alias.AliasType)
	}
	if alias.AliasName != "myhandle" {
		t.Errorf("Expected AliasName = 'myhandle', got '%s'", alias.AliasName)
	}
	if alias.OwnerDomainID != ownerID {
		t.Error("Expected owner to match ownerID")
	}
	if alias.TargetDomainID != targetID {
		t.Error("Expected target to match targetID")
	}
	if alias.GetDisplayName() != "My Handle" {
		t.Errorf("Expected display_name = 'My Handle', got '%s'", alias.GetDisplayName())
	}
}

func TestCreateRelationshipAlias_NameConflict(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	aliasRepo := repo.NewAliasRepository(pool)
	service := NewAliasService(aliasRepo)

	ownerID, targetID := setupTestDomains(t, pool)

	// Create first alias
	alias1, err := service.CreateRelationshipAlias(ctx, ownerID, targetID, "conflictname", nil, nil)
	if err != nil {
		t.Fatalf("CreateRelationshipAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias1.ID)

	// Try to create duplicate (should fail)
	_, err = service.CreateRelationshipAlias(ctx, ownerID, targetID, "conflictname", nil, nil)
	if err == nil {
		t.Fatal("Expected CreateRelationshipAlias to fail with duplicate name, got nil error")
	}
}

func TestCreateRelationshipAlias_RequiresName(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	aliasRepo := repo.NewAliasRepository(pool)
	service := NewAliasService(aliasRepo)

	ownerID, targetID := setupTestDomains(t, pool)

	// Try to create without name (should fail)
	_, err := service.CreateRelationshipAlias(ctx, ownerID, targetID, "", nil, nil)
	if err == nil {
		t.Fatal("Expected CreateRelationshipAlias to fail without name, got nil error")
	}
}

func TestCreateMaskAlias_AutoGenerateName(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	aliasRepo := repo.NewAliasRepository(pool)
	service := NewAliasService(aliasRepo)

	ownerID, _ := setupTestDomains(t, pool)

	// Create MASK without providing name
	alias, err := service.CreateMaskAlias(ctx, ownerID, "", 0, nil)
	if err != nil {
		t.Fatalf("CreateMaskAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias.ID)

	// Verify it's a MASK alias
	if alias.AliasType != domain.AliasTypeMask {
		t.Errorf("Expected AliasType = MASK, got %s", alias.AliasType)
	}

	// Verify name was generated
	if alias.AliasName == "" {
		t.Error("Expected auto-generated name, got empty string")
	}
	if len(alias.AliasName) < 5 || alias.AliasName[:5] != "mask-" {
		t.Errorf("Expected name starting with 'mask-', got '%s'", alias.AliasName)
	}

	// Verify self-targeting
	if alias.OwnerDomainID != ownerID {
		t.Error("Expected owner to match ownerID")
	}
	if alias.TargetDomainID != ownerID {
		t.Error("Expected target to match ownerID (self-targeting)")
	}
}

func TestCreateMaskAlias_WithTTL(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	aliasRepo := repo.NewAliasRepository(pool)
	service := NewAliasService(aliasRepo)

	ownerID, _ := setupTestDomains(t, pool)

	// Create MASK with TTL
	alias, err := service.CreateMaskAlias(ctx, ownerID, "custom-mask", 30, nil)
	if err != nil {
		t.Fatalf("CreateMaskAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias.ID)

	// Verify TTL was stored
	if alias.GetTTLDays() != 30 {
		t.Errorf("Expected TTL = 30 days, got %d", alias.GetTTLDays())
	}

	// Verify custom name was used
	if alias.AliasName != "custom-mask" {
		t.Errorf("Expected AliasName = 'custom-mask', got '%s'", alias.AliasName)
	}
}

func TestCreateMaskAlias_WithMetadata(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	aliasRepo := repo.NewAliasRepository(pool)
	service := NewAliasService(aliasRepo)

	ownerID, _ := setupTestDomains(t, pool)

	metadata := domain.AliasMetadata{
		"purpose": "browsing",
		"notes":   "test persona",
	}

	alias, err := service.CreateMaskAlias(ctx, ownerID, "", 7, metadata)
	if err != nil {
		t.Fatalf("CreateMaskAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias.ID)

	// Verify metadata was preserved and TTL added
	if alias.GetTTLDays() != 7 {
		t.Errorf("Expected TTL = 7 days, got %d", alias.GetTTLDays())
	}
	if alias.Metadata["purpose"] != "browsing" {
		t.Errorf("Expected purpose = 'browsing', got '%v'", alias.Metadata["purpose"])
	}
	if alias.Metadata["notes"] != "test persona" {
		t.Errorf("Expected notes = 'test persona', got '%v'", alias.Metadata["notes"])
	}
}

func TestRetireAlias(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	aliasRepo := repo.NewAliasRepository(pool)
	service := NewAliasService(aliasRepo)

	ownerID, targetID := setupTestDomains(t, pool)

	// Create alias
	alias, err := service.CreateRelationshipAlias(ctx, ownerID, targetID, "retire-test", nil, nil)
	if err != nil {
		t.Fatalf("CreateRelationshipAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias.ID)

	// Retire alias
	err = service.RetireAlias(ctx, alias.ID, "user request")
	if err != nil {
		t.Fatalf("RetireAlias failed: %v", err)
	}

	// Verify alias is retired
	retrieved, err := aliasRepo.GetAliasByID(ctx, alias.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve alias: %v", err)
	}

	if retrieved.RetiredAt == nil {
		t.Error("Expected RetiredAt to be set")
	}
	if retrieved.Status() != domain.AliasStatusRetired {
		t.Errorf("Expected status = RETIRED, got %s", retrieved.Status())
	}
}

func TestGetAliasesForDomain(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	aliasRepo := repo.NewAliasRepository(pool)
	service := NewAliasService(aliasRepo)

	ownerID, targetID := setupTestDomains(t, pool)

	// Create aliases where ownerID is owner
	alias1, err := service.CreateRelationshipAlias(ctx, ownerID, targetID, "alias1", nil, nil)
	if err != nil {
		t.Fatalf("CreateRelationshipAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias1.ID)

	// Create alias where ownerID is target
	alias2, err := service.CreateRelationshipAlias(ctx, targetID, ownerID, "alias2", nil, nil)
	if err != nil {
		t.Fatalf("CreateRelationshipAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias2.ID)

	// Get all aliases for ownerID (both as owner and target)
	aliases, err := service.GetAliasesForDomain(ctx, ownerID, false)
	if err != nil {
		t.Fatalf("GetAliasesForDomain failed: %v", err)
	}

	// Should find at least both aliases
	foundAsOwner := false
	foundAsTarget := false
	for _, a := range aliases {
		if a.ID == alias1.ID {
			foundAsOwner = true
		}
		if a.ID == alias2.ID {
			foundAsTarget = true
		}
	}

	if !foundAsOwner {
		t.Error("Did not find alias where ownerID is owner")
	}
	if !foundAsTarget {
		t.Error("Did not find alias where ownerID is target")
	}
}

func TestGetAliasesForDomain_Deduplication(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	aliasRepo := repo.NewAliasRepository(pool)
	service := NewAliasService(aliasRepo)

	ownerID, _ := setupTestDomains(t, pool)

	// Create self-targeting alias (owner = target)
	alias, err := service.CreateMaskAlias(ctx, ownerID, "selftest", 0, nil)
	if err != nil {
		t.Fatalf("CreateMaskAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias.ID)

	// Get aliases for ownerID
	aliases, err := service.GetAliasesForDomain(ctx, ownerID, false)
	if err != nil {
		t.Fatalf("GetAliasesForDomain failed: %v", err)
	}

	// Count occurrences of self-targeting alias (should be deduplicated)
	count := 0
	for _, a := range aliases {
		if a.ID == alias.ID {
			count++
		}
	}

	if count != 1 {
		t.Errorf("Expected self-targeting alias to appear once, got %d times", count)
	}
}
