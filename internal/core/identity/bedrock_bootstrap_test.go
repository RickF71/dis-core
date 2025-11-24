package identity_test

import (
	"context"
	"testing"

	"dis-core/internal/core/bedrock"
	"dis-core/internal/core/identity"
	"dis-core/internal/testdb"

	"github.com/google/uuid"
)

func TestBedrockCreatesFullSpine(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback(ctx)

	if err := bedrock.BedrockBootstrapAtomic(ctx, tx); err != nil {
		t.Fatalf("bedrock bootstrap failed: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// verify domains exist
	var cnt int
	if err := pool.QueryRow(ctx, `SELECT COUNT(1) FROM domains WHERE name IN ('null','void','terra','numen','lima')`).Scan(&cnt); err != nil {
		t.Fatalf("query domains: %v", err)
	}
	if cnt < 5 {
		t.Fatalf("expected 5 spine domains, got %d", cnt)
	}
}

func TestKnowThyselfFirstLogin_PerformsBedrockBootstrap(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()
	// create handshake for alice
	token := uuid.NewString()
	subject := "alice"
	if _, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject, status) VALUES ($1,$2,$3,$4)`, uuid.NewString(), token, subject, ""); err != nil {
		t.Fatalf("insert handshake: %v", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback(ctx)

	presentationName := "Alice"
	res, err := identity.KnowThyselfAtomic(ctx, tx, token, presentationName)
	if err != nil {
		t.Fatalf("KnowThyselfAtomic: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// verify system_state bootstrapped
	var val string
	if err := pool.QueryRow(ctx, `SELECT value FROM system_state WHERE key = 'bootstrapped' LIMIT 1`).Scan(&val); err != nil {
		t.Fatalf("query system_state: %v", err)
	}
	if val != "1" {
		t.Fatalf("expected system bootstrapped flag set, got %q", val)
	}

	// verify corporeal domain created with the human-facing presentation name
	domainName := presentationName
	var domainID string
	if err := pool.QueryRow(ctx, `SELECT id FROM domains WHERE name = $1 LIMIT 1`, domainName).Scan(&domainID); err != nil {
		t.Fatalf("expected corporeal domain for subject: %v", err)
	}

	// verify prime seat exists for actor/domain
	var seatCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(1) FROM domain_seats WHERE domain_id = $1::uuid AND COALESCE(member_id::text,'') = $2 AND seat_type = 'prime'`, domainID, res.ActorID.String()).Scan(&seatCount); err != nil {
		t.Fatalf("query domain_seats: %v", err)
	}
	if seatCount != 1 {
		t.Fatalf("expected 1 prime seat for actor/domain, got %d", seatCount)
	}
}

func TestSystemBootstrappedSkipsBedrockOnSecondLogin(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()
	// first user
	token1 := uuid.NewString()
	subject1 := "bob"
	if _, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject, status) VALUES ($1,$2,$3,$4)`, uuid.NewString(), token1, subject1, ""); err != nil {
		t.Fatalf("insert handshake1: %v", err)
	}
	tx1, _ := pool.Begin(ctx)
	res1, err := identity.KnowThyselfAtomic(ctx, tx1, token1, "Bob")
	if err != nil {
		t.Fatalf("first KnowThyselfAtomic: %v", err)
	}
	if err := tx1.Commit(ctx); err != nil {
		t.Fatalf("commit1: %v", err)
	}

	// second user should not re-run bedrock but should create their identity/domain
	token2 := uuid.NewString()
	subject2 := "carol"
	if _, err := pool.Exec(ctx, `INSERT INTO handshakes (id, token, subject, status) VALUES ($1,$2,$3,$4)`, uuid.NewString(), token2, subject2, ""); err != nil {
		t.Fatalf("insert handshake2: %v", err)
	}
	tx2, _ := pool.Begin(ctx)
	res2, err := identity.KnowThyselfAtomic(ctx, tx2, token2, "Carol")
	if err != nil {
		t.Fatalf("second KnowThyselfAtomic: %v", err)
	}
	if err := tx2.Commit(ctx); err != nil {
		t.Fatalf("commit2: %v", err)
	}

	if res1 == nil || res2 == nil {
		t.Fatalf("unexpected nil result(s)")
	}
}
