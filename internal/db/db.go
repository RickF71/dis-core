package db

import (
	"database/sql"
	"dis-core/internal/canon"
	"fmt"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var DefaultDB *sql.DB // ✅ global handle for daemon & status API

// SetupDatabase opens (and if needed creates) the Postgres database and ensures all core tables exist.
func SetupDatabase() (*sql.DB, error) {
	// Prefer environment variable for flexibility
	dsn := os.Getenv("DIS_DB_DSN")
	if dsn == "" {
		// Default local fallback
		dsn = "postgres://dis_user:card567@localhost:5432/dis_core?sslmode=disable"
		fmt.Println("⚠️  Using default Postgres DSN:", dsn)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	// Sanity check connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	// --- Ensure core tables ---
	if err := EnsureReceiptsSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure receipts: %w", err)
	}
	if err := EnsureIdentitiesSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure identities: %w", err)
	}
	if err := EnsureHandshakesSchema(db); err != nil {
		fmt.Println("⚠️  Handshakes schema not created:", err)
	}

	if err := EnsurePeersSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure peers schema: %w", err)
	}

	// --- Ensure peers table ---
	if err := EnsurePeersSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure peers schema: %w", err)
	}

	// --- Self-bootstrapping Domain schema ---
	if err := EnsureDomainSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("ensure domain schema: %w", err)
	}

	// --- Seed base domains ---
	if err := SeedDefaultDomains(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("seed default domains: %w", err)
	}

	// --- Seed base domains ---
	if err := SeedDefaultDomains(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("seed default domains: %w", err)
	}

	// --- Self-canonize domains ---
	if err := canon.ExportDomains(db, "domains/_auto"); err != nil {
		fmt.Printf("⚠️  Canon export failed: %v\n", err)
	} else {
		fmt.Println("✅ Canonical domain export complete.")
	}

	DefaultDB = db
	fmt.Println("✅ PostgreSQL database ready:", dsn)
	return db, nil
}

// CloseDatabase safely closes DefaultDB (optional helper)
func CloseDatabase() {
	if DefaultDB != nil {
		_ = DefaultDB.Close()
		DefaultDB = nil
	}
}

//
// ──────────────────────────────────────────────
//   Self-Bootstrapping Domain Schema + Seed
// ──────────────────────────────────────────────
//

// EnsurePeersSchema creates the peers table if it does not exist.
func EnsurePeersSchema(db *sql.DB) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS peers (
		address TEXT PRIMARY KEY,
		last_seen TIMESTAMPTZ DEFAULT NOW(),
		healthy BOOLEAN DEFAULT FALSE
	);
	`
	if _, err := db.Exec(stmt); err != nil {
		return fmt.Errorf("create peers table: %w", err)
	}
	fmt.Println("✅ peers table ready.")
	return nil
}

// EnsureDomainSchema creates the domains table if it does not exist.
func EnsureDomainSchema(db *sql.DB) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS domains (
		id TEXT PRIMARY KEY,
		parent_id TEXT REFERENCES domains(id),
		name TEXT NOT NULL,
		is_notech BOOLEAN NOT NULL DEFAULT FALSE,
		requires_inside_domain BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`
	_, err := db.Exec(stmt)
	if err != nil {
		return fmt.Errorf("create domains table: %w", err)
	}
	return nil
}

// Canonize performs a one-time export of all domains into canonical YAML files.
func Canonize(db *sql.DB) {
	if err := canon.ExportDomains(db, "domains/_auto"); err != nil {
		fmt.Printf("⚠️ Canon export failed: %v\n", err)
	} else {
		fmt.Println("✅ Canonical domain export complete.")
	}
}

// SeedDefaultDomains inserts the base DIS domains if they don't exist.
func SeedDefaultDomains(db *sql.DB) error {
	baseDomains := []struct {
		ID       string
		ParentID *string
		Name     string
		IsNotech bool
	}{
		{"domain.null", nil, "NULL", false},
		{"domain.terra", strPtr("domain.null"), "Terra", false},
	}

	for _, d := range baseDomains {
		var exists bool
		err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM domains WHERE id = $1)`, d.ID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check %s: %w", d.ID, err)
		}
		if !exists {
			_, err := db.Exec(`
				INSERT INTO domains (id, parent_id, name, is_notech, requires_inside_domain)
				VALUES ($1, $2, $3, $4, TRUE)`,
				d.ID, d.ParentID, d.Name, d.IsNotech)
			if err != nil {
				return fmt.Errorf("seed %s: %w", d.ID, err)
			}
			fmt.Printf("🌱 Seeded base domain: %s\n", d.ID)
		}
	}
	return nil
}

func EnsureChairsSchema(db *sql.DB) error {
	stmts := []string{
		// Core chair definition
		`CREATE TABLE IF NOT EXISTS chairs (
			id TEXT PRIMARY KEY,
			domain_id TEXT NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			requires_inside_domain BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,

		// Mapping of current occupant(s)
		`CREATE TABLE IF NOT EXISTS chairholders (
			chair_id TEXT NOT NULL REFERENCES chairs(id) ON DELETE CASCADE,
			identity_id TEXT NOT NULL,
			authorized_until TIMESTAMPTZ,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (chair_id, identity_id)
		);`,

		// Simple enforcement: cannot hold a chair outside your domain if requires_inside_domain=true
		`CREATE OR REPLACE FUNCTION enforce_inside_domain()
		RETURNS TRIGGER AS $$
		BEGIN
			IF (SELECT requires_inside_domain FROM chairs WHERE id = NEW.chair_id) THEN
				-- ensure occupant belongs to same domain
				PERFORM 1 FROM identities
					WHERE id = NEW.identity_id
					  AND domain_id = (SELECT domain_id FROM chairs WHERE id = NEW.chair_id);
				IF NOT FOUND THEN
					RAISE EXCEPTION 'Chairholder % cannot sit in chair %: not InsideDomain', NEW.identity_id, NEW.chair_id;
				END IF;
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;`,

		`CREATE OR REPLACE TRIGGER trg_enforce_inside_domain
			BEFORE INSERT OR UPDATE ON chairholders
			FOR EACH ROW
			EXECUTE FUNCTION enforce_inside_domain();`,
	}

	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("ensure chairs schema: %w", err)
		}
	}

	// ✅ just return the error from seeding (not two values)
	if err := SeedDefaultChairs(db); err != nil {
		return fmt.Errorf("seed default chairs: %w", err)
	}

	return nil
}

func SeedDefaultChairs(db *sql.DB) error {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM chairs WHERE id = 'chair.integration')`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check chairs: %w", err)
	}
	if !exists {
		_, err := db.Exec(`
			INSERT INTO chairs (id, domain_id, title, requires_inside_domain)
			VALUES ('chair.integration', 'domain.terra', 'Integration Chair', TRUE)
		`)
		if err != nil {
			return fmt.Errorf("seed chairs: %w", err)
		}
		fmt.Println("🌱 Seeded base chair: chair.integration (domain.terra)")
	}
	return nil
}

// helper
func strPtr(s string) *string { return &s }
