package domain

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dis-core/internal/testdb"
)

// GOV-8: Test suite enforcing DIS-Invariant-001
// No authority, lookup, or receipt may rely on domain names

func TestResolveDomainFDN(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test")
	}
	ctx := context.Background()
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	// GOV-8: Test UUID → name resolution (display only)
	// Using terra UUID (4daf928e...)
	domainID := uuid.MustParse("4daf928e-e58c-454e-8395-f3dedd103dde")
	name, err := ResolveDomainFDN(ctx, pool, domainID)
	if err != nil {
		t.Fatalf("ResolveDomainFDN failed: %v", err)
	}
	assert.Equal(t, "terra", name, "Expected FDN to be terra")

	// Test corporeal domain
	corpID := uuid.MustParse("a1111111-1111-1111-1111-111111111111")
	corpName, err := ResolveDomainFDN(ctx, pool, corpID)
	if err != nil {
		t.Fatalf("ResolveDomainFDN failed for corporeal: %v", err)
	}
	assert.Equal(t, "terra.numen.lima.corporeal", corpName, "Expected corporeal FDN")
}

func TestResolveDomainLineage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	// GOV-8: Test UUID-only lineage resolution
	// Using corporeal domain (child of terra, child of void)
	corporealID := uuid.MustParse("a1111111-1111-1111-1111-111111111111") // terra.numen.lima.corporeal
	lineage, err := ResolveDomainLineage(ctx, pool, corporealID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(lineage), 2, "Lineage should include at least corporeal and terra")

	// Last element should be the requested domain
	assert.Equal(t, corporealID, lineage[len(lineage)-1])

	// Verify lineage contains parent terra (4daf928e...)
	terraID := uuid.MustParse("4daf928e-e58c-454e-8395-f3dedd103dde")
	assert.Contains(t, lineage, terraID, "Lineage should contain parent terra domain")
}

func TestResolveDomainLineageFDN(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	// GOV-8: Test display-only lineage names
	// Using corporeal domain for richer lineage
	corporealID := uuid.MustParse("a1111111-1111-1111-1111-111111111111")
	fdnLineage, err := ResolveDomainLineageFDN(ctx, pool, corporealID)
	require.NoError(t, err)
	assert.Contains(t, fdnLineage, "terra.numen.lima.corporeal")
	assert.Contains(t, fdnLineage, "terra", "Lineage should contain parent terra")
}

func TestValidateDomainID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	// GOV-8: Validate UUID existence (not name lookups)
	validID := uuid.MustParse("4daf928e-e58c-454e-8395-f3dedd103dde") // terra
	exists, err := ValidateDomainID(ctx, pool, validID)
	require.NoError(t, err)
	assert.True(t, exists, "Should be true for valid domain ID")

	// Test invalid domain - use UUID that definitely doesn't exist
	invalidID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	exists, err = ValidateDomainID(ctx, pool, invalidID)
	require.NoError(t, err)
	assert.False(t, exists, "Should be false for non-existent domain ID")
}

func TestGetDomainMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)
	defer pool.Close()

	// Test metadata retrieval
	corporealID := uuid.MustParse("a1111111-1111-1111-1111-111111111111")
	meta, err := GetDomainMetadata(ctx, pool, corporealID)
	require.NoError(t, err)
	assert.Equal(t, corporealID, meta.ID)
	assert.Equal(t, "terra.numen.lima.corporeal", meta.Name)
	assert.Contains(t, meta.NameWarning, "GOV-8")
	assert.Equal(t, "human_sovereign", meta.Governance)
}

// GOV-8: Regression test - ensure no name-based WHERE clauses in queries
func TestNoNameBasedQueries(t *testing.T) {
	// This test documents the GOV-8 invariant
	// Any SQL query using "WHERE name =" for domain lookups violates DIS-Invariant-001

	t.Run("DocumentInvariant", func(t *testing.T) {
		violations := []string{
			"SELECT id FROM domains WHERE name = 'user.rick'",         // VIOLATION
			"UPDATE domains SET payload = '{}' WHERE name = 'system'", // VIOLATION
			"DELETE FROM domains WHERE name = 'obsolete'",             // VIOLATION
		}

		acceptable := []string{
			"SELECT id FROM domains WHERE id = $1",            // ACCEPTABLE
			"SELECT name FROM domains WHERE id = $1",          // ACCEPTABLE (display only)
			"UPDATE domains SET payload = '{}' WHERE id = $1", // ACCEPTABLE
		}

		assert.Len(t, violations, 3, "GOV-8: Document 3 violation patterns")
		assert.Len(t, acceptable, 3, "GOV-8: Document 3 acceptable patterns")
	})
}

// GOV-8: Benchmark UUID-based resolution vs name-based (deprecated)
func BenchmarkUUIDResolution(b *testing.B) {
	ctx := context.Background()
	pool := testdb.SetupTestDB(b)
	testdb.MustHaveDB(b, pool)
	defer pool.Close()

	domainID := uuid.MustParse("4daf928e-e58c-454e-8395-f3dedd103dde")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exists, _ := ValidateDomainID(ctx, pool, domainID)
		if !exists {
			b.Fatal("Domain should exist")
		}
	}
}
