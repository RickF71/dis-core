package testdb

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"

	"dis-core/internal/bootstrap"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// setupOnce ensures the destructive schema reset + bootstrap runs only
	// once per test process. Tests call SetupTestDB frequently; doing the
	// DROP/CREATE for each call can cause other concurrently-running tests
	// in the same process to observe missing relations if a subsequent
	// setup runs while a test is executing. Using sync.Once avoids that.
	setupOnce sync.Once
	testPool  *pgxpool.Pool
	setupErr  error
)

// adminLockPool holds an admin connection that keeps the advisory lock for
// the lifetime of the test process. We intentionally do not Close() it so
// the advisory lock remains held and prevents other test processes from
// dropping the shared dis_test schema while this process runs its tests.
var adminLockPool *pgxpool.Pool
var setupMu sync.Mutex

// SetupTestDB tries to create a pgxpool.Pool based on DIS_TEST_DB_DSN.
// Accepts testing.TB so it can be used from both tests and benchmarks.
// If the env var is not set, it returns nil and logs a note.
func SetupTestDB(tb testing.TB) *pgxpool.Pool {
	tb.Helper()

	dsn := os.Getenv("DIS_TEST_DB_DSN")
	if dsn == "" {
		tb.Log("DIS_TEST_DB_DSN not set; running tests without a real DB pool")
		return nil
	}

	// Enable synchronous receipt recording for tests by default. This makes
	// background receipt-writing deterministic for the harness and avoids
	// races where goroutines write after pools are closed or tables are
	// truncated. Tests may override by unsetting or changing the env var.
	// (Option A)
	_ = os.Setenv("DIS_TEST_SYNC_RECEIPTS", "1")
	// Allow idempotent test-only domain upserts used to defend against
	// FK races when background writers are active during tests. Gate this
	// so production behavior is not changed by the defensive upsert.
	_ = os.Setenv("DIS_TEST_ALLOW_DOMAIN_UPSERT", "1")

	// Run the destructive reset + bootstrap exactly once per process while
	// holding a long-lived admin connection that retains the advisory lock.
	// This prevents other test processes from mutating the shared dis_test
	// schema while this process runs its tests. We keep the admin pool open
	// for the lifetime of the process and return fresh pools to callers so
	// they may Close() them without affecting the admin lock.
	setupOnce.Do(func() {
		ctx := context.Background()
		const advisoryLockID = 4242424242424242

		// Create admin lock pool
		ap, err := pgxpool.New(ctx, dsn)
		if err != nil {
			setupErr = err
			return
		}

		// Acquire advisory lock and intentionally DO NOT release it until
		// process exit (we keep ap referenced in adminLockPool).
		if _, err := ap.Exec(ctx, `SELECT pg_advisory_lock($1)`, advisoryLockID); err != nil {
			ap.Close()
			setupErr = err
			return
		}

		// Reset public schema once per process
		if _, err := ap.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
			ap.Exec(ctx, `SELECT pg_advisory_unlock($1)`, advisoryLockID)
			ap.Close()
			setupErr = err
			return
		}

		// Ensure pgcrypto
		if _, err := ap.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pgcrypto;`); err != nil {
			ap.Exec(ctx, `SELECT pg_advisory_unlock($1)`, advisoryLockID)
			ap.Close()
			setupErr = err
			return
		}

		// Optionally apply on-disk migrations
		if os.Getenv("DIS_TEST_APPLY_MIGRATIONS") == "1" {
			// locate migrations dir (best-effort upward search)
			cwd, _ := os.Getwd()
			dir := cwd
			migrationsDir := ""
			for i := 0; i < 12; i++ {
				candidate := filepath.Join(dir, "db", "migrations")
				if info, err := os.Stat(candidate); err == nil && info.IsDir() {
					migrationsDir = candidate
					break
				}
				parent := filepath.Dir(dir)
				if parent == dir {
					break
				}
				dir = parent
			}
			if migrationsDir != "" {
				entries, err := os.ReadDir(migrationsDir)
				if err == nil {
					names := make([]string, 0, len(entries))
					for _, e := range entries {
						if e.IsDir() {
							continue
						}
						names = append(names, e.Name())
					}
					sort.Strings(names)
					for _, name := range names {
						if filepath.Ext(name) != ".sql" {
							continue
						}
						path := filepath.Join(migrationsDir, name)
						sqlBytes, err := os.ReadFile(path)
						if err != nil {
							ap.Exec(ctx, `SELECT pg_advisory_unlock($1)`, advisoryLockID)
							ap.Close()
							setupErr = err
							return
						}
						if _, err := ap.Exec(ctx, string(sqlBytes)); err != nil {
							ap.Exec(ctx, `SELECT pg_advisory_unlock($1)`, advisoryLockID)
							ap.Close()
							setupErr = err
							return
						}
					}
				}
			}
		}

		// Programmatic bootstrap
		if err := bootstrap.BootstrapAllTables(ap); err != nil {
			ap.Exec(ctx, `SELECT pg_advisory_unlock($1)`, advisoryLockID)
			ap.Close()
			setupErr = err
			return
		}

		// Save admin pool (keep open) so advisory lock is held for process lifetime
		adminLockPool = ap
	})

	if setupErr != nil {
		tb.Fatalf("failed to initialize test DB: %v", setupErr)
	}

	// Return a fresh pool per caller — callers may Close() this pool.
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		tb.Fatalf("failed to create test DB pool (DIS_TEST_DB_DSN=%q): %v", dsn, err)
	}

	// Ensure the test database is clean for this caller by truncating all
	// public tables. We serialize truncation in-process to avoid races.
	setupMu.Lock()
	defer setupMu.Unlock()

	// Use the adminLockPool (if available) to perform truncation so the
	// advisory lock holder executes the destructive cleanup.
	execPool := pool
	if adminLockPool != nil {
		execPool = adminLockPool
	}

	ctx := context.Background()
	// Gather public tables
	rows, err := execPool.Query(ctx, `SELECT tablename FROM pg_tables WHERE schemaname='public'`)
	if err != nil {
		pool.Close()
		tb.Fatalf("failed to list tables for cleanup: %v", err)
	}

	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			rows.Close()
			pool.Close()
			tb.Fatalf("failed to scan table name: %v", err)
		}
		// Skip the special pg_catalog tables (shouldn't be present here) and
		// skip possible migration bookkeeping if present.
		if t == "schema_migrations" {
			continue
		}
		tables = append(tables, t)
	}
	rows.Close()

	if len(tables) > 0 {
		// Build TRUNCATE statement
		stmt := "TRUNCATE TABLE "
		for i, t := range tables {
			if i > 0 {
				stmt += ", "
			}
			stmt += "public." + t
		}
		stmt += " RESTART IDENTITY CASCADE;"

		if _, err := execPool.Exec(ctx, stmt); err != nil {
			pool.Close()
			tb.Fatalf("failed to truncate test tables: %v", err)
		}
	}

	// Seed canonical domains expected by GOV-8 tests. We insert after truncation
	// so each caller receives the same deterministic domain fixtures.
	seedInserts := []string{
		// Insert aether as child of null when present, then ensure terra is parented under aether.
		"INSERT INTO domains (id, name, parent_id, domain_type, payload, created_at, updated_at) VALUES (gen_random_uuid(), 'aether', (SELECT id FROM domains WHERE name='null' OR name='domain.null' LIMIT 1), 'aether', '{}'::jsonb, now(), now()) ON CONFLICT (name) DO NOTHING",
		"INSERT INTO domains (id, name, parent_id, domain_type, payload, created_at, updated_at) VALUES ('4daf928e-e58c-454e-8395-f3dedd103dde', 'terra', (SELECT id FROM domains WHERE name='aether' OR name='domain.aether' LIMIT 1), 'terra', '{\"meta\": {\"governance\": \"human_sovereign\"}}'::jsonb, now(), now())",
		"INSERT INTO domains (id, name, parent_id, domain_type, payload, created_at, updated_at) VALUES ('a1111111-1111-1111-1111-111111111111', 'terra.numen.lima.corporeal', (SELECT id FROM domains WHERE name='terra' OR name='domain.terra' LIMIT 1), 'corporeal', '{\"meta\": {\"governance\": \"human_sovereign\"}}'::jsonb, now(), now())",
	}

	for _, s := range seedInserts {
		if _, err := execPool.Exec(ctx, s); err != nil {
			pool.Close()
			tb.Fatalf("failed to seed test domains: %v", err)
		}
	}

	return pool
}

// MustHaveDB skips the current test/benchmark if the pool is nil.
func MustHaveDB(tb testing.TB, pool *pgxpool.Pool) {
	tb.Helper()
	if pool == nil {
		tb.Skip("Skipping test: DIS_TEST_DB_DSN not set or DB unavailable")
	}
}

// SeedCanonicalSpine inserts the canonical spine domains in order (void, null, aether, terra, numen, lima, corporeal).
// It is idempotent and uses ON CONFLICT DO NOTHING so tests can call it repeatedly.
func SeedCanonicalSpine(pool *pgxpool.Pool) error {
	ctx := context.Background()
	// Insert in order so parent lookups succeed
	inserts := []struct {
		Name   string
		Parent string // empty means NULL
		Type   string
	}{
		{Name: "void", Parent: "", Type: "void"},
		{Name: "null", Parent: "void", Type: "null"},
		{Name: "aether", Parent: "null", Type: "aether"},
		{Name: "terra", Parent: "aether", Type: "terra"},
		{Name: "numen", Parent: "terra", Type: "numen"},
		{Name: "lima", Parent: "numen", Type: "lima"},
		{Name: "corporeal", Parent: "lima", Type: "corporeal"},
	}

	for _, it := range inserts {
		if it.Parent == "" {
			if _, err := pool.Exec(ctx, `INSERT INTO domains (id, name, parent_id, domain_type, payload, created_at, updated_at) VALUES (gen_random_uuid(), $1, NULL, $2, '{}'::jsonb, now(), now()) ON CONFLICT (name) DO NOTHING`, it.Name, it.Type); err != nil {
				return err
			}
		} else {
			if _, err := pool.Exec(ctx, `INSERT INTO domains (id, name, parent_id, domain_type, payload, created_at, updated_at) VALUES (gen_random_uuid(), $1, (SELECT id FROM domains WHERE name = $2 OR name = ('domain.'||$2) LIMIT 1), $3, '{}'::jsonb, now(), now()) ON CONFLICT (name) DO NOTHING`, it.Name, it.Parent, it.Type); err != nil {
				return err
			}
		}
	}
	// Ensure parent_id relationships are fixed for any existing rows that
	// were inserted earlier without parents (seed order or prior seeds).
	for _, it := range inserts {
		if it.Parent == "" {
			continue
		}
		if _, err := pool.Exec(ctx, `
			UPDATE domains SET parent_id = (
				SELECT id FROM domains WHERE name = $1 OR name = ('domain.'||$1) LIMIT 1
			)
			WHERE name = $2 AND (parent_id IS NULL)
		`, it.Parent, it.Name); err != nil {
			return err
		}
	}

	return nil
}
