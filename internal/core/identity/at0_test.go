package identity_test

import (
	"context"
	"testing"

	"dis-core/internal/core/identity"
	"dis-core/internal/testdb"

	"github.com/google/uuid"
)

// First identity should trigger AT-0: null domain + null.pseat created and
// bound to the new identity.
func TestAT0_FirstIdentityTriggersNullPseat(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()
	token := "at0-token-1"
	id := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject, status) VALUES ($1,$2,'', 'genesis')`, id, token); err != nil {
		t.Fatalf("insert handshake: %v", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer tx.Rollback(ctx)

	who := "Null First"
	res, err := identity.KnowThyselfAtomic(ctx, tx, token, who)
	if err != nil {
		t.Fatalf("KnowThyselfAtomic failed: %v", err)
	}

	if res.PseatID == uuid.Nil {
		t.Fatalf("expected PseatID to be set for AT-0; got nil")
	}

	// Commit so other sessions see the seat
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Verify the seat exists and is assigned to the actor
	var seatMember string
	if err := pool.QueryRow(ctx, `SELECT COALESCE(member_id::text,'') FROM domain_seats WHERE id = $1::uuid`, res.PseatID.String()).Scan(&seatMember); err != nil {
		t.Fatalf("query seat: %v", err)
	}
	if seatMember != res.ActorID.String() {
		t.Fatalf("seat not assigned to actor: got=%s want=%s", seatMember, res.ActorID.String())
	}
}

// Subsequent identities should NOT trigger AT-0.
func TestAT0_SubsequentIdentityDoesNotTrigger(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()
	// first identity
	token1 := "at0-token-2"
	id1 := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject, status) VALUES ($1,$2,'', 'genesis')`, id1, token1); err != nil {
		t.Fatalf("insert handshake: %v", err)
	}
	tx1, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx1: %v", err)
	}
	_, err = identity.KnowThyselfAtomic(ctx, tx1, token1, "First Actor")
	if err != nil {
		t.Fatalf("first KnowThyselfAtomic failed: %v", err)
	}
	if err := tx1.Commit(ctx); err != nil {
		t.Fatalf("commit tx1: %v", err)
	}

	// second identity
	token2 := "at0-token-3"
	id2 := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id2, token2, "human:bob"); err != nil {
		t.Fatalf("insert handshake2: %v", err)
	}
	tx2, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx2: %v", err)
	}
	res2, err := identity.KnowThyselfAtomic(ctx, tx2, token2, "Bob")
	if err != nil {
		t.Fatalf("second KnowThyselfAtomic failed: %v", err)
	}
	if err := tx2.Commit(ctx); err != nil {
		t.Fatalf("commit tx2: %v", err)
	}

	if res2.PseatID != uuid.Nil {
		t.Fatalf("expected subsequent identity NOT to receive a null pseat; got PseatID=%s", res2.PseatID.String())
	}
}

// null.pseat should be marked immutable via seat_roles
func TestAT0_NullPseatImmutable(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()
	token := "at0-token-4"
	id := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject, status) VALUES ($1,$2,'', 'genesis')`, id, token); err != nil {
		t.Fatalf("insert handshake: %v", err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	res, err := identity.KnowThyselfAtomic(ctx, tx, token, "Immutable Actor")
	if err != nil {
		t.Fatalf("KnowThyselfAtomic failed: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// verify seat_roles contains immutable role for pseat
	var gotRole string
	if err := pool.QueryRow(ctx, `SELECT role FROM seat_roles WHERE seat_id = $1::uuid AND role = 'immutable' LIMIT 1`, res.PseatID.String()).Scan(&gotRole); err != nil {
		t.Fatalf("query seat_roles: %v", err)
	}
	if gotRole != "immutable" {
		t.Fatalf("expected immutable role on pseat; got=%s", gotRole)
	}
}
