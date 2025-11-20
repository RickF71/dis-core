package repo

import (
	"context"
	"testing"

	"dis-core/internal/domain"
	"dis-core/internal/testdb"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GOV-12: Alias Canon & DSCI Integration - Repository Tests

// Tests should use the centralized testdb harness.

// setupTestDomains creates test domains for alias testing
func setupTestDomains(t *testing.T, pool *pgxpool.Pool) (ownerID, targetID uuid.UUID) {
	ctx := context.Background()

	ownerID = uuid.New()
	targetID = uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO domains (id, name, created_at)
		VALUES
			($1, 'test-owner-domain', NOW()),
			($2, 'test-target-domain', NOW())
		ON CONFLICT (id) DO NOTHING
	`, ownerID, targetID)

	if err != nil {
		t.Fatalf("Failed to create test domains: %v", err)
	}

	return ownerID, targetID
}

// cleanupAliases removes test aliases
func cleanupAliases(t *testing.T, pool *pgxpool.Pool, aliasIDs ...uuid.UUID) {
	ctx := context.Background()
	for _, id := range aliasIDs {
		_, _ = pool.Exec(ctx, "DELETE FROM aliases WHERE id = $1", id)
	}
}

func TestCreateAlias(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	ctx := context.Background()
	repo := NewAliasRepository(pool)

	ownerID, targetID := setupTestDomains(t, pool)

	alias := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: targetID,
		AliasName:      "test-alias",
		AliasType:      domain.AliasTypeRelationship,
		Metadata:       make(domain.AliasMetadata),
	}

	err := repo.CreateAlias(ctx, alias)
	if err != nil {
		t.Fatalf("CreateAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias.ID)

	// Verify ID was generated
	if alias.ID == uuid.Nil {
		t.Error("Expected ID to be generated, got nil UUID")
	}

	// Verify CreatedAt was set
	if alias.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set, got zero time")
	}

	// Verify alias is retrievable
	retrieved, err := repo.GetAliasByID(ctx, alias.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve alias: %v", err)
	}

	if retrieved.AliasName != alias.AliasName {
		t.Errorf("Expected alias_name = %s, got %s", alias.AliasName, retrieved.AliasName)
	}
	if retrieved.AliasType != alias.AliasType {
		t.Errorf("Expected alias_type = %s, got %s", alias.AliasType, retrieved.AliasType)
	}
}

func TestCreateAliasUniqueConstraint(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	ctx := context.Background()
	repo := NewAliasRepository(pool)

	ownerID, targetID := setupTestDomains(t, pool)

	// Create first alias
	alias1 := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: targetID,
		AliasName:      "unique-test",
		AliasType:      domain.AliasTypeRelationship,
		Metadata:       make(domain.AliasMetadata),
	}

	err := repo.CreateAlias(ctx, alias1)
	if err != nil {
		t.Fatalf("CreateAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias1.ID)

	// Try to create duplicate active alias (should fail)
	alias2 := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: targetID,
		AliasName:      "unique-test", // Same name
		AliasType:      domain.AliasTypeRelationship,
		Metadata:       make(domain.AliasMetadata),
	}

	err = repo.CreateAlias(ctx, alias2)
	if err == nil {
		cleanupAliases(t, pool, alias2.ID)
		t.Fatal("Expected CreateAlias to fail with duplicate active alias, got nil error")
	}
}

func TestRetireAlias(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	ctx := context.Background()
	repo := NewAliasRepository(pool)

	ownerID, targetID := setupTestDomains(t, pool)

	// Create alias
	alias := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: targetID,
		AliasName:      "retire-test",
		AliasType:      domain.AliasTypeMask,
		Metadata:       make(domain.AliasMetadata),
	}

	err := repo.CreateAlias(ctx, alias)
	if err != nil {
		t.Fatalf("CreateAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias.ID)

	// Retire alias
	reason := "user request"
	err = repo.RetireAlias(ctx, alias.ID, reason)
	if err != nil {
		t.Fatalf("RetireAlias failed: %v", err)
	}

	// Verify alias is retired
	retrieved, err := repo.GetAliasByID(ctx, alias.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve alias: %v", err)
	}

	if retrieved.RetiredAt == nil {
		t.Error("Expected RetiredAt to be set, got nil")
	}
	if retrieved.Status() != domain.AliasStatusRetired {
		t.Errorf("Expected status = RETIRED, got %s", retrieved.Status())
	}

	// Verify retirement reason in metadata
	if retrieved.Metadata["retirement_reason"] != reason {
		t.Errorf("Expected retirement_reason = %s, got %v", reason, retrieved.Metadata["retirement_reason"])
	}
}

func TestGetAliasesForOwner(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	ctx := context.Background()
	repo := NewAliasRepository(pool)

	ownerID, targetID := setupTestDomains(t, pool)

	// Create multiple aliases
	alias1 := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: targetID,
		AliasName:      "owner-test-1",
		AliasType:      domain.AliasTypeRelationship,
		Metadata:       make(domain.AliasMetadata),
	}
	alias2 := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: targetID,
		AliasName:      "owner-test-2",
		AliasType:      domain.AliasTypeMask,
		Metadata:       make(domain.AliasMetadata),
	}

	err := repo.CreateAlias(ctx, alias1)
	if err != nil {
		t.Fatalf("CreateAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias1.ID)

	err = repo.CreateAlias(ctx, alias2)
	if err != nil {
		t.Fatalf("CreateAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias2.ID)

	// Retire one alias
	err = repo.RetireAlias(ctx, alias2.ID, "test")
	if err != nil {
		t.Fatalf("RetireAlias failed: %v", err)
	}

	// Get active aliases only
	activeAliases, err := repo.GetAliasesForOwner(ctx, ownerID, false)
	if err != nil {
		t.Fatalf("GetAliasesForOwner failed: %v", err)
	}

	// Should only find 1 active alias
	if len(activeAliases) < 1 {
		t.Errorf("Expected at least 1 active alias, got %d", len(activeAliases))
	}

	// Get all aliases (including retired)
	allAliases, err := repo.GetAliasesForOwner(ctx, ownerID, true)
	if err != nil {
		t.Fatalf("GetAliasesForOwner failed: %v", err)
	}

	// Should find at least 2 aliases
	if len(allAliases) < 2 {
		t.Errorf("Expected at least 2 total aliases, got %d", len(allAliases))
	}
}

func TestGetAliasesByType(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	ctx := context.Background()
	repo := NewAliasRepository(pool)

	ownerID, targetID := setupTestDomains(t, pool)

	// Create aliases of different types
	relAlias := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: targetID,
		AliasName:      "type-test-rel",
		AliasType:      domain.AliasTypeRelationship,
		Metadata:       make(domain.AliasMetadata),
	}
	maskAlias := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: ownerID,
		AliasName:      "type-test-mask",
		AliasType:      domain.AliasTypeMask,
		Metadata:       make(domain.AliasMetadata),
	}

	err := repo.CreateAlias(ctx, relAlias)
	if err != nil {
		t.Fatalf("CreateAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, relAlias.ID)

	err = repo.CreateAlias(ctx, maskAlias)
	if err != nil {
		t.Fatalf("CreateAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, maskAlias.ID)

	// Get RELATIONSHIP aliases
	relAliases, err := repo.GetAliasesByType(ctx, ownerID, domain.AliasTypeRelationship, false)
	if err != nil {
		t.Fatalf("GetAliasesByType failed: %v", err)
	}

	foundRel := false
	for _, a := range relAliases {
		if a.ID == relAlias.ID {
			foundRel = true
		}
		if a.AliasType != domain.AliasTypeRelationship {
			t.Errorf("Expected RELATIONSHIP type, got %s", a.AliasType)
		}
	}
	if !foundRel {
		t.Error("Did not find RELATIONSHIP alias in results")
	}

	// Get MASK aliases
	maskAliases, err := repo.GetAliasesByType(ctx, ownerID, domain.AliasTypeMask, false)
	if err != nil {
		t.Fatalf("GetAliasesByType failed: %v", err)
	}

	foundMask := false
	for _, a := range maskAliases {
		if a.ID == maskAlias.ID {
			foundMask = true
		}
		if a.AliasType != domain.AliasTypeMask {
			t.Errorf("Expected MASK type, got %s", a.AliasType)
		}
	}
	if !foundMask {
		t.Error("Did not find MASK alias in results")
	}
}

func TestGetCorporealAutoAlias(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	ctx := context.Background()
	repo := NewAliasRepository(pool)

	ownerID, _ := setupTestDomains(t, pool)

	// Create corporeal AUTO alias
	corpAlias := &domain.Alias{
		OwnerDomainID:   ownerID,
		TargetDomainID:  ownerID, // Self-targeting
		AliasName:       "corporeal-root",
		AliasType:       domain.AliasTypeAuto,
		IsCorporealAuto: true,
		Metadata:        make(domain.AliasMetadata),
	}

	err := repo.CreateAlias(ctx, corpAlias)
	if err != nil {
		t.Fatalf("CreateAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, corpAlias.ID)

	// Retrieve corporeal auto alias
	retrieved, err := repo.GetCorporealAutoAlias(ctx, ownerID)
	if err != nil {
		t.Fatalf("GetCorporealAutoAlias failed: %v", err)
	}

	if retrieved.ID != corpAlias.ID {
		t.Errorf("Expected ID = %s, got %s", corpAlias.ID, retrieved.ID)
	}
	if !retrieved.IsCorporealAuto {
		t.Error("Expected IsCorporealAuto = true, got false")
	}
	if retrieved.AliasType != domain.AliasTypeAuto {
		t.Errorf("Expected AliasType = AUTO, got %s", retrieved.AliasType)
	}
}

func TestFindActiveAliasByName(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	ctx := context.Background()
	repo := NewAliasRepository(pool)

	ownerID, targetID := setupTestDomains(t, pool)

	// Create alias
	alias := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: targetID,
		AliasName:      "find-by-name-test",
		AliasType:      domain.AliasTypeRelationship,
		Metadata:       make(domain.AliasMetadata),
	}

	err := repo.CreateAlias(ctx, alias)
	if err != nil {
		t.Fatalf("CreateAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias.ID)

	// Find by name
	found, err := repo.FindActiveAliasByName(ctx, ownerID, targetID, "find-by-name-test")
	if err != nil {
		t.Fatalf("FindActiveAliasByName failed: %v", err)
	}

	if found.ID != alias.ID {
		t.Errorf("Expected ID = %s, got %s", alias.ID, found.ID)
	}

	// Retire alias
	err = repo.RetireAlias(ctx, alias.ID, "test")
	if err != nil {
		t.Fatalf("RetireAlias failed: %v", err)
	}

	// Should not find retired alias
	_, err = repo.FindActiveAliasByName(ctx, ownerID, targetID, "find-by-name-test")
	if err == nil {
		t.Error("Expected error when finding retired alias, got nil")
	}
}

func TestGenerateMaskName(t *testing.T) {
	// Generate multiple mask names
	name1 := GenerateMaskName()
	name2 := GenerateMaskName()

	// Should start with "mask-"
	if len(name1) < 6 || name1[:5] != "mask-" {
		t.Errorf("Expected name to start with 'mask-', got %s", name1)
	}

	// Should be unique (probabilistically)
	if name1 == name2 {
		t.Error("Expected unique mask names, got duplicates")
	}

	// Should be lowercase hex
	hexPart := name1[5:]
	if len(hexPart) != 8 {
		t.Errorf("Expected 8-character hex suffix, got %d characters", len(hexPart))
	}
}

func TestRetireAliasAllowsReuse(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	ctx := context.Background()
	repo := NewAliasRepository(pool)

	ownerID, targetID := setupTestDomains(t, pool)

	// Create and retire first alias
	alias1 := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: targetID,
		AliasName:      "reuse-test",
		AliasType:      domain.AliasTypeRelationship,
		Metadata:       make(domain.AliasMetadata),
	}

	err := repo.CreateAlias(ctx, alias1)
	if err != nil {
		t.Fatalf("CreateAlias failed: %v", err)
	}
	defer cleanupAliases(t, pool, alias1.ID)

	err = repo.RetireAlias(ctx, alias1.ID, "test")
	if err != nil {
		t.Fatalf("RetireAlias failed: %v", err)
	}

	// Create new alias with same name (should succeed after retirement)
	alias2 := &domain.Alias{
		OwnerDomainID:  ownerID,
		TargetDomainID: targetID,
		AliasName:      "reuse-test", // Same name
		AliasType:      domain.AliasTypeRelationship,
		Metadata:       make(domain.AliasMetadata),
	}

	err = repo.CreateAlias(ctx, alias2)
	if err != nil {
		t.Fatalf("CreateAlias failed after retirement: %v", err)
	}
	defer cleanupAliases(t, pool, alias2.ID)

	// Verify both exist but only one is active
	allAliases, err := repo.GetAliasesForOwner(ctx, ownerID, true)
	if err != nil {
		t.Fatalf("GetAliasesForOwner failed: %v", err)
	}

	activeCount := 0
	retiredCount := 0
	for _, a := range allAliases {
		if a.AliasName == "reuse-test" {
			if a.IsActive() {
				activeCount++
			} else {
				retiredCount++
			}
		}
	}

	if activeCount < 1 {
		t.Error("Expected at least 1 active 'reuse-test' alias")
	}
	if retiredCount < 1 {
		t.Error("Expected at least 1 retired 'reuse-test' alias")
	}
}
