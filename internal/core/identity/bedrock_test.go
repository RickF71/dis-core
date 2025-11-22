package identity_test

import (
	"context"
	"testing"

	"dis-core/internal/core/identity"
	"dis-core/internal/discapsule"
	"dis-core/internal/testdb"
)

func TestKnowThyselfBedrock_CreatesNull(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback(ctx)

	grant := discapsule.BedrockGrant{Approved: true}
	if err := identity.KnowThyselfBedrock(ctx, tx, grant); err != nil {
		t.Fatalf("KnowThyselfBedrock failed: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// verify domain 'null' exists
	var cnt int
	if err := pool.QueryRow(ctx, `SELECT COUNT(1) FROM domains WHERE name = 'null'`).Scan(&cnt); err != nil {
		t.Fatalf("query domains: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 domain 'null', got %d", cnt)
	}
}

func TestKnowThyselfBedrock_Idempotent(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()
	// First run
	tx1, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx1: %v", err)
	}
	defer tx1.Rollback(ctx)
	grant := discapsule.BedrockGrant{Approved: true}
	if err := identity.KnowThyselfBedrock(ctx, tx1, grant); err != nil {
		t.Fatalf("first KnowThyselfBedrock failed: %v", err)
	}
	if err := tx1.Commit(ctx); err != nil {
		t.Fatalf("commit tx1: %v", err)
	}

	// Second run should be a no-op
	tx2, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx2: %v", err)
	}
	defer tx2.Rollback(ctx)
	if err := identity.KnowThyselfBedrock(ctx, tx2, grant); err != nil {
		t.Fatalf("second KnowThyselfBedrock failed: %v", err)
	}
	if err := tx2.Commit(ctx); err != nil {
		t.Fatalf("commit tx2: %v", err)
	}

	var cnt int
	if err := pool.QueryRow(ctx, `SELECT COUNT(1) FROM domains WHERE name = 'null'`).Scan(&cnt); err != nil {
		t.Fatalf("query domains: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected 1 domain 'null' after idempotent runs, got %d", cnt)
	}
}

func TestKnowThyselfBedrock_RejectsUnapprovedGrant(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback(ctx)

	grant := discapsule.BedrockGrant{Approved: false}
	if err := identity.KnowThyselfBedrock(ctx, tx, grant); err == nil {
		t.Fatalf("expected error for unapproved grant")
	}
}
