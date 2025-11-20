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

// setupTestDB connects to local Postgres test DB or skips the test.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	return pool
}

// setupTestDomain ensures there is at least one domain to use for seats.
func setupTestDomain(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	ctx := context.Background()

	var domainID uuid.UUID
	err := pool.QueryRow(ctx, "SELECT id FROM domains LIMIT 1").Scan(&domainID)
	if err == nil {
		return domainID
	}

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

// TestSeatReceiptChaining ensures instantiate + status create chained receipts
func TestSeatReceiptChaining(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	domainID := setupTestDomain(t, pool)

	engine := NewEngine(nil, pool)

	// Identity / active user
	identityID := uuid.NewString()
	au := &authpkg.ActiveUser{
		CorporealDomainUID: identityID,
		HasExternalUID:     true,
		Bound:              true,
		CorporealDomainID:  1,
	}
	ctx = authpkg.WithActiveUser(ctx, au)

	// Instantiate a seat
	seat, err := engine.InstantiateSeat(ctx, domainID.String(), identityID)
	if err != nil {
		t.Fatalf("InstantiateSeat failed: %v", err)
	}

	// Read receipts for seat
	r1, err := dbstore.ListSeatReceipts(ctx, pool, seat.ID)
	if err != nil {
		t.Fatalf("ListSeatReceipts failed: %v", err)
	}
	if len(r1) == 0 {
		t.Fatalf("expected at least one receipt after instantiate, got 0")
	}

	// Ensure first receipt has a hash in its payload
	firstPayload, ok := r1[0]["payload"].(map[string]interface{})
	if !ok {
		t.Fatalf("first receipt payload not present or wrong type")
	}
	h1, ok := firstPayload["hash"].(string)
	if !ok || h1 == "" {
		t.Fatalf("first receipt missing hash")
	}

	// Trigger a view (status) which records a second receipt (best-effort)
	_, err = engine.SeatStatus(ctx, seat.ID)
	if err != nil {
		t.Fatalf("SeatStatus failed: %v", err)
	}

	// Small sleep to ensure ordering
	time.Sleep(10 * time.Millisecond)

	r2, err := dbstore.ListSeatReceipts(ctx, pool, seat.ID)
	if err != nil {
		t.Fatalf("ListSeatReceipts after status failed: %v", err)
	}
	if len(r2) < 2 {
		t.Fatalf("expected >=2 receipts after status, got %d", len(r2))
	}

	// Examine last two receipts for chaining
	last := r2[len(r2)-1]
	prev := r2[len(r2)-2]
	lastPayload, ok := last["payload"].(map[string]interface{})
	if !ok {
		t.Fatalf("last receipt payload missing")
	}
	prevPayload, ok := prev["payload"].(map[string]interface{})
	if !ok {
		t.Fatalf("prev receipt payload missing")
	}

	lastPrev, _ := lastPayload["prev_hash"].(string)
	prevHash, _ := prevPayload["hash"].(string)
	if lastPrev == "" || prevHash == "" {
		t.Fatalf("expected non-empty hash fields; got prev=%q prevHash=%q", lastPrev, prevHash)
	}
	if lastPrev != prevHash {
		t.Fatalf("hash chain mismatch: last.prev_hash (%s) != prev.hash (%s)", lastPrev, prevHash)
	}

	// Cleanup
	_, _ = pool.Exec(ctx, "DELETE FROM receipts WHERE target = $1", seat.ID)
	_, _ = pool.Exec(ctx, "DELETE FROM domain_seats WHERE id = $1", seat.ID)
}

// TestInstantiateSeatAtomicity verifies that if a receipt write fails during
// an atomic create flow the seat insert is rolled back and no receipt exists.
func TestInstantiateSeatAtomicity(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	domainID := setupTestDomain(t, pool)

	// Start a transaction and simulate the instantiate flow manually:
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	seatID := uuid.NewString()
	identityID := uuid.NewString()

	// Insert seat inside tx
	if err := dbstore.InsertSeatTx(ctx, tx, seatID, domainID.String(), identityID, "member"); err != nil {
		tx.Rollback(ctx)
		t.Fatalf("InsertSeatTx failed: %v", err)
	}

	// Create a payload that cannot be marshaled to force RecordSeatReceiptTx to fail
	badPayload := map[string]any{"bad": make(chan int)}

	engine := NewEngine(nil, pool)
	// Attempt to record receipt (should fail)
	if err := engine.RecordSeatReceiptTx(ctx, tx, seatID, "seat.instantiate.v1", badPayload); err == nil {
		t.Fatalf("expected RecordSeatReceiptTx to fail for bad payload")
	}

	// Rollback and assert seat does not exist
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	// Ensure seat absent
	srow, serr := dbstore.GetSeat(ctx, pool, seatID)
	if serr != nil {
		t.Fatalf("GetSeat failed: %v", serr)
	}
	if srow != nil {
		t.Fatalf("expected seat to be absent after rollback, but found: %v", srow)
	}
}
