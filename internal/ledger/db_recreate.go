package ledger

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/lib/pq"
)

// RecreateDatabase drops the target DB (if exists), ensures the app role exists with the
// given password, creates the DB owned by that role, and sets sane default privileges.
// adminDSN: superuser/createdb role DSN (e.g. postgres://postgres:admin@localhost/postgres?sslmode=disable)
// dbName:   e.g. "dis_core"
// appUser:  e.g. "dis_user"
// appPass:  password for appUser
func RecreateDatabase(adminDSN, dbName, appUser, appPass string) error {
	// Always connect to the control DB
	adminDSN = withDBName(adminDSN, "postgres")

	admin, err := sql.Open("postgres", adminDSN)
	if err != nil {
		return fmt.Errorf("admin connect failed: %w", err)
	}
	defer admin.Close()

	// Terminate connections if they exist, ignore errors
	_, _ = admin.Exec(fmt.Sprintf(`
		REVOKE CONNECT ON DATABASE %s FROM PUBLIC;
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = '%s' AND pid <> pg_backend_pid();
	`, dbName, dbName))

	// Drop if it exists (ignore errors if already gone)
	_, _ = admin.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", pgIdent(dbName)))

	// Ensure app role exists and set password safely
	roleSQL := fmt.Sprintf(`
	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '%s') THEN
			CREATE ROLE %s WITH LOGIN PASSWORD '%s';
		ELSE
			ALTER ROLE %s WITH LOGIN PASSWORD '%s';
		END IF;
	END$$;
	`, appUser, pgIdent(appUser), appPass, pgIdent(appUser), appPass)

	if _, err := admin.Exec(roleSQL); err != nil {
		return fmt.Errorf("ensure role failed: %w", err)
	}

	// Create the database owned by appUser
	createDB := fmt.Sprintf(`
		CREATE DATABASE %s
		WITH OWNER %s
		ENCODING 'UTF8'
		LC_COLLATE='en_US.UTF-8'
		LC_CTYPE='en_US.UTF-8'
		TEMPLATE template0;
	`, pgIdent(dbName), pgIdent(appUser))
	if _, err := admin.Exec(createDB); err != nil {
		return fmt.Errorf("create database failed: %w", err)
	}

	// Connect to the newly created DB to apply ownership and privileges
	appDBdsn := withDBName(adminDSN, dbName)
	db, err := sql.Open("postgres", appDBdsn)
	if err != nil {
		return fmt.Errorf("post-create connect failed: %w", err)
	}
	defer db.Close()

	alters := []string{
		fmt.Sprintf("ALTER SCHEMA public OWNER TO %s;", pgIdent(appUser)),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s;", pgIdent(dbName), pgIdent(appUser)),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO %s;", pgIdent(appUser)),
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO %s;", pgIdent(appUser)),
	}
	for _, q := range alters {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("privileges setup failed: %w", err)
		}
	}

	fmt.Printf("✅ Recreated database %s with owner %s.\n", dbName, appUser)
	return nil
}

// pgIdent minimally quotes identifiers that might contain uppercase or special chars.
// For simple lowercase names without special chars it returns as-is.
func pgIdent(s string) string {
	if s == "" {
		return `""`
	}
	// If strictly [a-z0-9_] return as-is; else double-quote.
	for _, r := range s {
		if !(r == '_' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
			return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
		}
	}
	return s
}

// withDBName replaces the path (db name) in a postgres DSN using net/url parsing.
func withDBName(dsn, db string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}
	// Path is like "/dbname"
	u.Path = "/" + db
	return u.String()
}
