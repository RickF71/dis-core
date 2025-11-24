package identity_test

import (
	"context"
	"testing"

	"dis-core/internal/core/identity"
	"dis-core/internal/testdb"

	"github.com/google/uuid"
)

func TestKnowThyselfAtomic_HappyPath(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()
	// create handshake token
	token := "test-token-123"
	subject := "human:alice"
	id := uuid.NewString()
	_, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id, token, subject)
	if err != nil {
		t.Fatalf("failed to insert handshake: %v", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback(ctx)

	who := "Alice Example"
	res, err := identity.KnowThyselfAtomic(ctx, tx, token, who)
	if err != nil {
		t.Fatalf("KnowThyselfAtomic failed: %v", err)
	}

	if res.ActorID == uuid.Nil || res.DomainID == uuid.Nil || res.ReceiptID == uuid.Nil {
		t.Fatalf("invalid result ids: %+v", res)
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Stronger assertions: verify identity, seat, and receipt records exist and contain expected values
	// 1) identities.presentation_name should equal who
	var gotPresentation string
	if err := pool.QueryRow(ctx, `SELECT presentation_name FROM identities WHERE id = $1::uuid`, res.ActorID.String()).Scan(&gotPresentation); err != nil {
		t.Fatalf("failed to query identity presentation_name: %v", err)
	}
	if gotPresentation != who {
		t.Fatalf("presentation_name mismatch: got=%q want=%q", gotPresentation, who)
	}

	// 2) domain_seats should include a prime seat for actor/domain
	var seatID, seatDomain, seatMember, seatType string
	if err := pool.QueryRow(ctx, `SELECT id::text, domain_id::text, COALESCE(member_id::text,''), seat_type FROM domain_seats WHERE domain_id = $1::uuid AND COALESCE(member_id::text,'') = $2 AND seat_type = 'prime' LIMIT 1`, res.DomainID.String(), res.ActorID.String()).Scan(&seatID, &seatDomain, &seatMember, &seatType); err != nil {
		t.Fatalf("failed to query domain_seats for prime seat: %v", err)
	}
	if seatDomain != res.DomainID.String() || seatMember != res.ActorID.String() || seatType != "prime" {
		t.Fatalf("domain_seats row mismatch: domain=%s member=%s type=%s", seatDomain, seatMember, seatType)
	}

	// 3) receipts should contain the recorded login receipt with expected fields
	var rid, rtype string
	var payload map[string]interface{}
	if err := pool.QueryRow(ctx, `SELECT id::text, type, payload FROM receipts WHERE id = $1::uuid`, res.ReceiptID.String()).Scan(&rid, &rtype, &payload); err != nil {
		t.Fatalf("failed to query receipt by id: %v", err)
	}
	if rtype != "ci.call.login.v1" {
		t.Fatalf("receipt type mismatch: got=%s want=ci.call.login.v1", rtype)
	}
	// payload should include actor, domain, and presentation
	if payload == nil {
		t.Fatalf("receipt payload is nil")
	}
	if payload["actor"] != res.ActorID.String() {
		t.Fatalf("receipt payload actor mismatch: got=%v want=%v", payload["actor"], res.ActorID.String())
	}
	if payload["domain"] != res.DomainID.String() {
		t.Fatalf("receipt payload domain mismatch: got=%v want=%v", payload["domain"], res.DomainID.String())
	}
	if payload["presentation"] != who {
		t.Fatalf("receipt payload presentation mismatch: got=%v want=%v", payload["presentation"], who)
	}

	// At genesis we expect an explicit empty roles array in the receipt payload
	if proles, ok := payload["roles"]; ok {
		if arr, ok := proles.([]interface{}); ok {
			if len(arr) != 0 {
				t.Fatalf("expected empty roles array at genesis; got=%v", arr)
			}
		} else {
			t.Fatalf("receipt payload roles not array: %T", proles)
		}
	} else {
		t.Fatalf("receipt payload missing roles field")
	}
}

func TestKnowThyselfAtomic_InvalidToken(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback(ctx)

	_, err = identity.KnowThyselfAtomic(ctx, tx, "nonexistent-token", "Nobody")
	if err == nil {
		t.Fatalf("expected error for invalid token")
	}
}

// First corporeal KnowThyselfAtomic invocation should establish the root sovereign
// on the null domain if none exists. The second invocation should NOT overwrite it.
func TestKnowThyselfAtomic_EstablishesRootSovereign(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()

	// first actor (will cause root sovereign establishment)
	token1 := "root-token-1"
	id1 := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject) VALUES ($1,$2,$3)`, id1, token1, "human:alice"); err != nil {
		t.Fatalf("insert handshake1: %v", err)
	}
	tx1, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx1: %v", err)
	}
	res1, err := identity.KnowThyselfAtomic(ctx, tx1, token1, "Alice")
	if err != nil {
		t.Fatalf("first KnowThyselfAtomic: %v", err)
	}
	if err := tx1.Commit(ctx); err != nil {
		t.Fatalf("commit tx1: %v", err)
	}

	// find null domain id
	var nullDomainID string
	if err := pool.QueryRow(ctx, `SELECT id::text FROM domains WHERE name = 'domain.null' LIMIT 1`).Scan(&nullDomainID); err != nil {
		t.Fatalf("query null domain: %v", err)
	}

	// find prime seat in null domain
	var rootSeatID, rootSeatMember string
	if err := pool.QueryRow(ctx, `SELECT id::text, COALESCE(member_id::text,'') FROM domain_seats WHERE domain_id = $1::uuid AND seat_type = 'prime' LIMIT 1`, nullDomainID).Scan(&rootSeatID, &rootSeatMember); err != nil {
		t.Fatalf("query null prime seat: %v", err)
	}
	if rootSeatMember != res1.ActorID.String() {
		t.Fatalf("root seat not assigned to first actor: got=%s want=%s", rootSeatMember, res1.ActorID.String())
	}

	// second actor should NOT overwrite root seat
	token2 := "root-token-2"
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
		t.Fatalf("second KnowThyselfAtomic: %v", err)
	}
	if err := tx2.Commit(ctx); err != nil {
		t.Fatalf("commit tx2: %v", err)
	}

	// re-check root seat ownership
	var finalSeatMember string
	if err := pool.QueryRow(ctx, `SELECT COALESCE(member_id::text,'') FROM domain_seats WHERE id = $1::uuid`, rootSeatID).Scan(&finalSeatMember); err != nil {
		t.Fatalf("re-query root seat: %v", err)
	}
	if finalSeatMember != rootSeatMember {
		t.Fatalf("root seat was overwritten: before=%s after=%s", rootSeatMember, finalSeatMember)
	}

	// Ensure second actor got its own prime seat in its corporeal domain
	if res2.DomainID == uuid.Nil || res2.ActorID == uuid.Nil {
		t.Fatalf("second actor result invalid: %+v", res2)
	}
}
