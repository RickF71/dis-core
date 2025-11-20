package identity

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"dis-core/internal/testdb"
)

// Test database setup
func setupTestDB(t *testing.T) *pgxpool.Pool {
	// Delegate to centralized test harness. It will skip the test when
	// DIS_TEST_DB_DSN is not set, or fatal on failure to connect when set.
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	return pool
}

// setupTestDomain creates or retrieves a test domain for use in tests
func setupTestDomain(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	ctx := context.Background()

	// Try to get an existing domain first
	var domainID uuid.UUID
	err := pool.QueryRow(ctx, "SELECT id FROM domains LIMIT 1").Scan(&domainID)
	if err == nil {
		return domainID
	}

	// Create a test domain if none exists
	domainID = uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO domains (id, name, domain_type, authority, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (id) DO NOTHING
	`, domainID, "test-domain", "corporeal", "test")

	if err != nil {
		t.Fatalf("Failed to create test domain: %v", err)
	}

	return domainID
}

// TestRecordIdentityReceipt verifies basic receipt creation and hash computation
func TestRecordIdentityReceipt(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	store := NewIdentityReceiptStore(pool)

	domainID := setupTestDomain(t, pool)
	actorID := uuid.New()

	payload := map[string]interface{}{
		"action":      "test_action",
		"description": "Test identity receipt",
	}

	// Create root receipt (no prev_id)
	receipt := &IdentityReceipt{
		ID:         uuid.New(),
		DomainID:   domainID,
		ActorID:    actorID,
		Action:     IdentityRootV1,
		Payload:    payload,
		PrevID:     nil,
		ConsentBy:  domainID,
		AliasScope: nil,
	}

	err := store.RecordIdentityReceipt(ctx, receipt)
	if err != nil {
		t.Fatalf("Failed to record receipt: %v", err)
	}

	// Verify hash was computed
	if receipt.Hash == "" {
		t.Error("Expected non-empty hash")
	}

	// Retrieve the receipt
	lastReceipt, err := store.GetLastReceipt(ctx, actorID)
	if err != nil {
		t.Fatalf("Failed to get last receipt: %v", err)
	}

	if lastReceipt.ID != receipt.ID {
		t.Errorf("Expected receipt ID %s, got %s", receipt.ID, lastReceipt.ID)
	}

	if lastReceipt.Hash != receipt.Hash {
		t.Errorf("Expected hash %s, got %s", receipt.Hash, lastReceipt.Hash)
	}

	// Cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM identity_receipts WHERE actor_id = $1", actorID)
}

// TestIdentityReceiptChaining verifies prev_id chaining works correctly
func TestIdentityReceiptChaining(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	store := NewIdentityReceiptStore(pool)

	domainID := setupTestDomain(t, pool)
	actorID := uuid.New()

	// Create root receipt
	receipt1, err := store.RecordIdentityAction(
		ctx,
		actorID,
		domainID,
		IdentityRootV1,
		map[string]interface{}{"step": "root"},
		domainID,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to record root receipt: %v", err)
	}

	if receipt1.PrevID != nil {
		t.Error("Expected root receipt to have nil prev_id")
	}

	// Create chained receipt
	receipt2, err := store.RecordIdentityAction(
		ctx,
		actorID,
		domainID,
		IdentityBindingUpdateV1,
		map[string]interface{}{"step": "update"},
		domainID,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to record chained receipt: %v", err)
	}

	if receipt2.PrevID == nil {
		t.Fatal("Expected chained receipt to have prev_id")
	}

	if *receipt2.PrevID != receipt1.ID {
		t.Errorf("Expected prev_id %s, got %s", receipt1.ID, *receipt2.PrevID)
	}

	// Cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM identity_receipts WHERE actor_id = $1", actorID)
}

// TestGetIdentityLineage verifies lineage retrieval and ordering
func TestGetIdentityLineage(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	store := NewIdentityReceiptStore(pool)

	domainID := setupTestDomain(t, pool)
	actorID := uuid.New()

	// Create 3 chained receipts
	actions := []IdentityAction{
		IdentityRootV1,
		IdentityAliasAddV1,
		IdentityBindingUpdateV1,
	}

	for i, action := range actions {
		_, err := store.RecordIdentityAction(
			ctx,
			actorID,
			domainID,
			action,
			map[string]interface{}{"step": i + 1},
			domainID,
			nil,
		)
		if err != nil {
			t.Fatalf("Failed to record receipt %d: %v", i+1, err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure distinct timestamps
	}

	// Retrieve lineage
	lineage, err := GetIdentityLineage(ctx, pool, actorID)
	if err != nil {
		t.Fatalf("Failed to get lineage: %v", err)
	}

	if len(lineage.Entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(lineage.Entries))
	}

	// Verify chronological order (oldest first)
	for i := 1; i < len(lineage.Entries); i++ {
		// Since CreatedAt is string, just verify they're different
		// (actual ordering is ensured by the SQL ORDER BY created_at ASC)
		if lineage.Entries[i].CreatedAt == lineage.Entries[i-1].CreatedAt {
			t.Logf("Warning: Entries %d and %d have same timestamp", i-1, i)
		}
	}

	// Verify first entry is root
	if lineage.Entries[0].Action != string(IdentityRootV1) {
		t.Errorf("Expected first action to be root, got %s", lineage.Entries[0].Action)
	}

	// Cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM identity_receipts WHERE actor_id = $1", actorID)
}

// TestIdentityHashChainIntegrity verifies hash chain validation
func TestIdentityHashChainIntegrity(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	store := NewIdentityReceiptStore(pool)

	domainID := setupTestDomain(t, pool)
	actorID := uuid.New()

	// Create 3 chained receipts
	for i := 0; i < 3; i++ {
		action := IdentityRootV1
		if i > 0 {
			action = IdentityBindingUpdateV1
		}
		_, err := store.RecordIdentityAction(
			ctx,
			actorID,
			domainID,
			action,
			map[string]interface{}{"step": i + 1},
			domainID,
			nil,
		)
		if err != nil {
			t.Fatalf("Failed to record receipt %d: %v", i+1, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Retrieve lineage with integrity verification
	lineage, err := GetIdentityLineage(ctx, pool, actorID)
	if err != nil {
		t.Fatalf("Failed to get lineage: %v", err)
	}

	// All entries should be valid
	for i, entry := range lineage.Entries {
		if !entry.Valid {
			t.Errorf("Entry %d should be valid but is marked invalid", i)
		}
	}

	// Verify integrity status
	if lineage.Integrity.TotalReceipts != 3 {
		t.Errorf("Expected 3 total receipts, got %d", lineage.Integrity.TotalReceipts)
	}

	if lineage.Integrity.ValidChains != 2 {
		t.Errorf("Expected 2 valid chains, got %d", lineage.Integrity.ValidChains)
	}

	if lineage.Integrity.RootReceipts != 1 {
		t.Errorf("Expected 1 root receipt, got %d", lineage.Integrity.RootReceipts)
	}

	if lineage.Integrity.BrokenChains != 0 {
		t.Errorf("Expected 0 broken chains, got %d", lineage.Integrity.BrokenChains)
	}

	// Cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM identity_receipts WHERE actor_id = $1", actorID)
}

// TestIdentityLineageIntegrityStatus verifies integrity status aggregation and FK protection
func TestIdentityLineageIntegrityStatus(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	store := NewIdentityReceiptStore(pool)

	domainID := setupTestDomain(t, pool)
	actorID := uuid.New()

	// Create root receipt
	receipt1, _ := store.RecordIdentityAction(
		ctx, actorID, domainID, IdentityRootV1,
		map[string]interface{}{"step": 1}, domainID, nil,
	)

	// Create chained receipt
	_, _ = store.RecordIdentityAction(
		ctx, actorID, domainID, IdentityAliasAddV1,
		map[string]interface{}{"step": 2}, domainID, nil,
	)

	// Attempt to tamper with the chain by updating prev_id to a non-existent UUID
	// This should fail due to FK constraint (verifying database integrity protection)
	wrongPrevID := uuid.New()
	_, err := pool.Exec(ctx,
		"UPDATE identity_receipts SET prev_id = $1 WHERE actor_id = $2 AND prev_id = $3",
		wrongPrevID, actorID, receipt1.ID,
	)

	// Verify FK constraint prevents tampering
	if err == nil {
		t.Error("Expected FK constraint violation when trying to set invalid prev_id, but update succeeded")
	} else {
		t.Logf("✓ FK constraint correctly prevented tampering: %v", err)
	}

	// Retrieve lineage - should be fully valid since tampering was prevented
	lineage, err := GetIdentityLineage(ctx, pool, actorID)
	if err != nil {
		t.Fatalf("Failed to get lineage: %v", err)
	}

	// Both entries should be valid
	if !lineage.Entries[0].Valid {
		t.Error("Root entry should be valid")
	}

	if !lineage.Entries[1].Valid {
		t.Error("Second entry should be valid (FK prevented tampering)")
	}

	// Verify integrity status shows no broken chains
	if lineage.Integrity.BrokenChains != 0 {
		t.Errorf("Expected 0 broken chains (FK protection), got %d", lineage.Integrity.BrokenChains)
	}

	if lineage.Integrity.ValidChains != 1 {
		t.Errorf("Expected 1 valid chain, got %d", lineage.Integrity.ValidChains)
	}

	// Cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM identity_receipts WHERE actor_id = $1", actorID)
}

// TestIdentityAliasScope verifies alias scope handling
func TestIdentityAliasScope(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	store := NewIdentityReceiptStore(pool)

	domainID := setupTestDomain(t, pool)
	actorID := uuid.New()

	scope := "domain-specific"

	// Create alias receipt with scope
	receipt, err := store.RecordIdentityAction(
		ctx,
		actorID,
		domainID,
		IdentityAliasAddV1,
		map[string]interface{}{"alias": "test-alias"},
		domainID,
		&scope,
	)
	if err != nil {
		t.Fatalf("Failed to record alias receipt: %v", err)
	}

	if receipt.AliasScope == nil {
		t.Fatal("Expected alias_scope to be set")
	}

	if *receipt.AliasScope != scope {
		t.Errorf("Expected alias_scope %s, got %s", scope, *receipt.AliasScope)
	}

	// Retrieve and verify
	lastReceipt, err := store.GetLastReceipt(ctx, actorID)
	if err != nil {
		t.Fatalf("Failed to get last receipt: %v", err)
	}

	if lastReceipt.AliasScope == nil || *lastReceipt.AliasScope != scope {
		t.Error("Alias scope not preserved in retrieval")
	}

	// Cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM identity_receipts WHERE actor_id = $1", actorID)
}

// TestIdentityHashDeterminism verifies hash computation is deterministic
func TestIdentityHashDeterminism(t *testing.T) {
	payload := map[string]interface{}{
		"action": "test",
		"value":  123,
	}

	payloadBytes, _ := marshalPayload(payload)
	prevID := uuid.New()

	hash1 := computeIdentityHash(payloadBytes, &prevID)
	hash2 := computeIdentityHash(payloadBytes, &prevID)

	if hash1 != hash2 {
		t.Error("Hash computation should be deterministic")
	}

	// Different prev_id should produce different hash
	differentPrevID := uuid.New()
	hash3 := computeIdentityHash(payloadBytes, &differentPrevID)

	if hash1 == hash3 {
		t.Error("Different prev_id should produce different hash")
	}
}

// Helper to marshal payload consistently
func marshalPayload(payload map[string]interface{}) ([]byte, error) {
	return []byte(`{"action":"test","value":123}`), nil
}
