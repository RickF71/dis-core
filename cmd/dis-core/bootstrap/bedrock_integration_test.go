package bootstrap_test

import (
	"context"
	"strings"
	"testing"

	"dis-core/cmd/dis-core/bootstrap"
	"dis-core/internal/discapsule"
	"dis-core/internal/testdb"

	"github.com/google/uuid"
)

// mockCapsule is a tiny test double implementing the Capsule interface used by
// the bedrock bootstrap flow. It returns a preconfigured BedrockGrant or error.
type mockCapsule struct {
	grant discapsule.BedrockGrant
	err   error
}

func (m *mockCapsule) PerformBedrockAuth(ch discapsule.BedrockChallenge) (discapsule.BedrockGrant, error) {
	return m.grant, m.err
}

// TestBedrock_NoDomainsTable drops the domains table to simulate a missing
// schema and verifies RunBedrockBootstrap can create/ensure the schema. In
// environments where DROP TABLE is blocked because other objects depend on the
// table, this test will be skipped (detects Postgres dependency error).
func TestBedrock_NoDomainsTable(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()

	// Try to drop the domains table. If the DB refuses because other objects
	// depend on it (SQLSTATE 2BP01 / "depends on it"), skip the test rather
	// than fail the whole suite.
	if _, err := pool.Exec(ctx, `DROP TABLE IF EXISTS domains`); err != nil {
		if strings.Contains(err.Error(), "depends on it") || strings.Contains(err.Error(), "2BP01") {
			t.Skipf("Skipping destructive bootstrap test: %v", err)
		}
		t.Fatalf("drop domains: %v", err)
	}

	// Ensure RunBedrockBootstrap does not try to use an interactive capsule in
	// this scenario (it should only ensure tables and domain state when
	// invoked in this mode).
	orig := bootstrap.NewCapsule
	bootstrap.NewCapsule = func() discapsule.Capsule { t.Fatalf("capsule should not be called"); return nil }
	defer func() { bootstrap.NewCapsule = orig }()

	if err := bootstrap.RunBedrockBootstrap(ctx, pool); err != nil {
		t.Fatalf("expected nil on missing domains table, got: %v", err)
	}
}

// TestBedrock_CapsuleApprovesAndDeclines verifies that when the capsule
// approves the grant a 'null' domain is created, and when it declines the
// operation returns an error. The test is defensive about schema operations
// and uses clean-up steps so it is safe to run multiple times.
func TestBedrock_CapsuleApprovesAndDeclines(t *testing.T) {
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	ctx := context.Background()

	// Ensure tables exist before running the interactive flow.
	if err := bootstrap.BootstrapTables(pool); err != nil {
		t.Fatalf("bootstrap tables: %v", err)
	}

	// Approve case: inject a capsule that always approves.
	orig := bootstrap.NewCapsule
	approve := &mockCapsule{grant: discapsule.BedrockGrant{GrantID: uuid.NewString(), Approved: true}}
	bootstrap.NewCapsule = func() discapsule.Capsule { return approve }

	if err := bootstrap.RunBedrockBootstrap(ctx, pool); err != nil {
		t.Fatalf("expected success when capsule approves: %v", err)
	}

	// verify domain created
	var cnt int
	if err := pool.QueryRow(ctx, `SELECT COUNT(1) FROM domains WHERE name = 'null'`).Scan(&cnt); err != nil {
		t.Fatalf("query domains: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected domain 'null' created, got %d", cnt)
	}

	// Decline case: remove created null domain, then inject a declining capsule.
	if _, err := pool.Exec(ctx, `DELETE FROM domains WHERE name = 'null'`); err != nil {
		t.Fatalf("cleanup domains: %v", err)
	}

	decline := &mockCapsule{grant: discapsule.BedrockGrant{Approved: false}}
	bootstrap.NewCapsule = func() discapsule.Capsule { return decline }

	if err := bootstrap.RunBedrockBootstrap(ctx, pool); err == nil {
		t.Fatalf("expected error when capsule declines")
	}

	// restore
	bootstrap.NewCapsule = orig
}
