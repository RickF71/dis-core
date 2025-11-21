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
