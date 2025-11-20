package authority

import (
	"context"
	"testing"
	"time"

	authpkg "dis-core/internal/auth"
	dbstore "dis-core/internal/db"
	"dis-core/internal/testdb"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// setupTestDB helper
func setupTestDBFreeze(t *testing.T) *pgxpool.Pool {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	return pool
}

// ensure there's at least one domain and return its UUID
func ensureTestDomain(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	ctx := context.Background()
	var domainID uuid.UUID
	if err := pool.QueryRow(ctx, "SELECT id FROM domains LIMIT 1").Scan(&domainID); err == nil {
		return domainID
	}
	domainID = uuid.New()
	_, err := pool.Exec(ctx, `
        INSERT INTO domains (id, name, domain_type, authority, created_at)
        VALUES ($1, $2, $3, $4, NOW())
        ON CONFLICT (id) DO NOTHING
    `, domainID, "freeze-test", "corporeal", "test")
	if err != nil {
		t.Fatalf("failed to create domain: %v", err)
	}
	return domainID
}

func TestFreezeDomain_HappyPath(t *testing.T) {
	pool := setupTestDBFreeze(t)
	defer pool.Close()

	ctx := context.Background()
	domainID := ensureTestDomain(t, pool)

	eng := NewEngine(nil, pool)

	// Active user for receipt actor
	au := &authpkg.ActiveUser{CorporealDomainUID: uuid.NewString()}
	ctx = authpkg.WithActiveUser(ctx, au)

	uid, err := eng.FreezeDomain(ctx, domainID, "all", "test freeze", uuid.New(), nil)
	if err != nil {
		t.Fatalf("FreezeDomain failed: %v", err)
	}
	if uid == uuid.Nil {
		t.Fatalf("expected non-nil freeze id")
	}

	// Ensure active freeze returned by store
	frs, err := dbstore.GetActiveFreezes(ctx, pool, domainID.String())
	if err != nil {
		t.Fatalf("GetActiveFreezes failed: %v", err)
	}
	if len(frs) == 0 {
		t.Fatalf("expected active freeze, got none")
	}

	// Ensure a receipt was written for the domain
	rows, err := pool.Query(ctx, `SELECT id FROM receipts WHERE domain::text = $1 OR (payload->>'domain_id') = $1`, domainID.String())
	if err != nil {
		t.Fatalf("query receipts failed: %v", err)
	}
	defer rows.Close()
	found := false
	for rows.Next() {
		found = true
		break
	}
	if !found {
		t.Fatalf("expected at least one receipt for domain after freeze")
	}
}

func TestUnfreezeDomain_HappyPath(t *testing.T) {
	pool := setupTestDBFreeze(t)
	defer pool.Close()

	ctx := context.Background()
	domainID := ensureTestDomain(t, pool)

	eng := NewEngine(nil, pool)
	// create freeze first
	_, err := eng.FreezeDomain(ctx, domainID, "all", "temp", uuid.New(), nil)
	if err != nil {
		t.Fatalf("setup FreezeDomain failed: %v", err)
	}

	if err := eng.UnfreezeDomain(ctx, domainID, "all", uuid.New(), "unfreeze"); err != nil {
		t.Fatalf("UnfreezeDomain failed: %v", err)
	}

	frs, err := dbstore.GetActiveFreezes(ctx, pool, domainID.String())
	if err != nil {
		t.Fatalf("GetActiveFreezes failed: %v", err)
	}
	if len(frs) != 0 {
		t.Fatalf("expected no active freezes after unfreeze, got %d", len(frs))
	}
}

func TestOverrideFreezeDomain(t *testing.T) {
	pool := setupTestDBFreeze(t)
	defer pool.Close()

	ctx := context.Background()
	domainID := ensureTestDomain(t, pool)

	eng := NewEngine(nil, pool)
	priorID, err := eng.FreezeDomain(ctx, domainID, "write", "prior", uuid.New(), nil)
	if err != nil {
		t.Fatalf("FreezeDomain prior failed: %v", err)
	}

	newID, err := eng.OverrideFreezeDomain(ctx, domainID, "write", uuid.New(), "override", priorID)
	if err != nil {
		t.Fatalf("OverrideFreezeDomain failed: %v", err)
	}
	if newID == uuid.Nil {
		t.Fatalf("expected override id")
	}

	frs, err := dbstore.GetActiveFreezes(ctx, pool, domainID.String())
	if err != nil {
		t.Fatalf("GetActiveFreezes failed: %v", err)
	}
	if len(frs) != 1 {
		t.Fatalf("expected one active override freeze, got %d", len(frs))
	}
	if frs[0].OverrideOf == nil || *frs[0].OverrideOf == "" {
		t.Fatalf("expected override_of to be set on new freeze")
	}
}

func TestFreezeDomain_AlreadyFrozen(t *testing.T) {
	pool := setupTestDBFreeze(t)
	defer pool.Close()

	ctx := context.Background()
	domainID := ensureTestDomain(t, pool)

	eng := NewEngine(nil, pool)
	_, err := eng.FreezeDomain(ctx, domainID, "events", "first", uuid.New(), nil)
	if err != nil {
		t.Fatalf("first freeze failed: %v", err)
	}
	// second freeze same scope should replace first
	_, err = eng.FreezeDomain(ctx, domainID, "events", "second", uuid.New(), nil)
	if err != nil {
		t.Fatalf("second freeze failed: %v", err)
	}
	frs, err := dbstore.GetActiveFreezes(ctx, pool, domainID.String())
	if err != nil {
		t.Fatalf("GetActiveFreezes failed: %v", err)
	}
	if len(frs) != 1 {
		t.Fatalf("expected single active freeze after replacing, got %d", len(frs))
	}
}

func TestUnfreezeDomain_NoFreezeExists(t *testing.T) {
	pool := setupTestDBFreeze(t)
	defer pool.Close()

	ctx := context.Background()
	domainID := ensureTestDomain(t, pool)

	eng := NewEngine(nil, pool)
	// Should not error even if none exists
	if err := eng.UnfreezeDomain(ctx, domainID, "policy", uuid.New(), "no-op"); err != nil {
		t.Fatalf("UnfreezeDomain returned error when none existed: %v", err)
	}
}

func TestFreezeDomain_Scoped(t *testing.T) {
	pool := setupTestDBFreeze(t)
	defer pool.Close()

	ctx := context.Background()
	domainID := ensureTestDomain(t, pool)

	eng := NewEngine(nil, pool)
	_, err := eng.FreezeDomain(ctx, domainID, "events", "scoped", uuid.New(), nil)
	if err != nil {
		t.Fatalf("FreezeDomain scoped failed: %v", err)
	}
	frs, err := dbstore.GetActiveFreezes(ctx, pool, domainID.String())
	if err != nil {
		t.Fatalf("GetActiveFreezes failed: %v", err)
	}
	if len(frs) != 1 || frs[0].Scope != "events" {
		t.Fatalf("expected one events-scoped freeze, got: %#v", frs)
	}
}

func TestFreezeDomain_TTLExpiration(t *testing.T) {
	pool := setupTestDBFreeze(t)
	defer pool.Close()

	ctx := context.Background()
	domainID := ensureTestDomain(t, pool)

	eng := NewEngine(nil, pool)
	// small TTL: now + 1s
	ttl := time.Now().UTC().Add(1 * time.Second)
	_, err := eng.FreezeDomain(ctx, domainID, "write", "shortlived", uuid.New(), &ttl)
	if err != nil {
		t.Fatalf("FreezeDomain with TTL failed: %v", err)
	}
	// immediately should be present
	frs, err := dbstore.GetActiveFreezes(ctx, pool, domainID.String())
	if err != nil {
		t.Fatalf("GetActiveFreezes failed: %v", err)
	}
	if len(frs) != 1 {
		t.Fatalf("expected active freeze immediately after insert")
	}
	// wait for expiration
	time.Sleep(1500 * time.Millisecond)
	frs2, err := dbstore.GetActiveFreezes(ctx, pool, domainID.String())
	if err != nil {
		t.Fatalf("GetActiveFreezes after TTL failed: %v", err)
	}
	if len(frs2) != 0 {
		t.Fatalf("expected no active freezes after TTL expiration, got %d", len(frs2))
	}
}

func TestFreezeDomain_ReceiptAndFreeze_Atomic(t *testing.T) {
	pool := setupTestDBFreeze(t)
	defer pool.Close()

	ctx := context.Background()
	domainID := ensureTestDomain(t, pool)

	// Begin manual tx to simulate atomic insertion + failed receipt
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx failed: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Insert freeze row
	fid, err := dbstore.InsertDomainFreezeTx(ctx, tx, domainID.String(), "write", "atomic-test", uuid.NewString(), nil, nil)
	if err != nil {
		t.Fatalf("InsertDomainFreezeTx failed: %v", err)
	}

	// Now attempt to record a receipt with an unserializable payload
	badPayload := map[string]any{"bad": make(chan int)}
	eng := NewEngine(nil, pool)
	if err := eng.RecordDomainReceiptTx(ctx, tx, domainID.String(), "domain.freeze.v1", badPayload); err == nil {
		t.Fatalf("expected RecordDomainReceiptTx to fail with bad payload")
	}

	// Rollback and assert freeze not visible
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}
	frs, err := dbstore.GetActiveFreezes(ctx, pool, domainID.String())
	if err != nil {
		t.Fatalf("GetActiveFreezes failed: %v", err)
	}
	for _, f := range frs {
		if f.ID == fid {
			t.Fatalf("found freeze inserted in rolled-back tx: %s", fid)
		}
	}
}
