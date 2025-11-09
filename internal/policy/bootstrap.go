package policy

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
)

type defaultPolicy struct {
	name    string
	content string
}

// EnsureNullPolicies checks for missing null-domain policies
// and inserts the defaults if needed.
func EnsureNullPolicies(ctx context.Context, db *sql.DB) error {
	defaults := []defaultPolicy{
		{
			name: "gates",
			content: `package gates

# Default deny unless explicitly allowed
default allow = false

# Example rule: allow override if authorized
allow {
  input.action_ref == "domain.freeze.override.v1"
  input.consent_status == "granted"
}`,
		},
		{
			name: "risk",
			content: `package risk

# Default risk score
default score = 0.0

score = 0.1 { input.action_ref == "domain.freeze.v1" }
score = 0.9 { input.action_ref == "domain.freeze.override.v1" }`,
		},
		{
			name: "freeze",
			content: `package freeze

# Default freeze state
default active = false

active {
  input.action_ref == "domain.freeze.v1"
}`,
		},
	}

	for _, p := range defaults {
		var count int
		err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM policies WHERE domain_id IS NULL AND name = $1`, p.name,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("check %s: %w", p.name, err)
		}

		if count == 0 {
			_, err := db.ExecContext(ctx,
				`INSERT INTO policies (domain_id, name, content) VALUES (NULL, $1, $2)`,
				p.name, p.content)
			if err != nil {
				return fmt.Errorf("insert %s: %w", p.name, err)
			}
			fmt.Printf("[policy] seeded null policy: %s\n", p.name)
		}
	}

	fmt.Println("[policy] null policies ensured (gates, risk, freeze)")
	return nil
}

// EnsureDefaultPolicies creates the default policies for DIS-Core bootstrap
// This function reads actual rego files from disk and loads them
func EnsureDefaultPolicies(db *pgxpool.Pool) error {
	log.Println("[policy-bootstrap] Starting default policy initialization...")

	// Define the policy files to load
	policyFiles := []string{
		"internal/policy/gates.rego",
		"internal/policy/risk.rego",
		"internal/policy/freeze.rego",
	}

	// Load each policy file
	policyBundles := make(map[string]string)

	for _, filePath := range policyFiles {
		fileName := filepath.Base(filePath)

		log.Printf("[policy-bootstrap] Loading %s...", fileName)

		content, err := LoadRegoBundle(filePath)
		if err != nil {
			log.Printf("[policy-bootstrap] Warning: Could not load %s: %v", fileName, err)
			continue
		}

		// Calculate file size for logging
		sizeKB := float64(len(content)) / 1024.0
		log.Printf("[policy-bootstrap] Loaded %s (%.1f KB)", fileName, sizeKB)

		// Store in bundle map
		policyName := fileName[:len(fileName)-len(filepath.Ext(fileName))] // Remove .rego extension
		policyBundles[policyName] = content

		// Optional: Insert into policies table if it exists
		if err := storePolicyInDatabase(db, policyName, content); err != nil {
			log.Printf("[policy-bootstrap] Warning: Could not store %s in database: %v", policyName, err)
			// Continue without failing - keep in memory
		}
	}

	log.Printf("[policy-bootstrap] Successfully loaded %d policy files", len(policyBundles))
	log.Println("[policy-bootstrap] Default null-domain policies ensured.")
	return nil
}

// LoadRegoBundle reads a rego file and returns its contents as a string
func LoadRegoBundle(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read rego file %s: %w", path, err)
	}
	return string(data), nil
}

// storePolicyInDatabase attempts to store policy content in the database
func storePolicyInDatabase(db *pgxpool.Pool, name, content string) error {
	// Check if policies table exists
	var tableExists bool
	err := db.QueryRow(context.Background(),
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'policies')").Scan(&tableExists)
	if err != nil || !tableExists {
		return fmt.Errorf("policies table does not exist")
	}

	// Insert or update policy
	_, err = db.Exec(context.Background(),
		`INSERT INTO policies (name, domain_id, content, created_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (name) DO UPDATE SET
		 content = EXCLUDED.content, updated_at = NOW()`,
		name, "00000000-0000-0000-0000-000000000000", content)

	return err
}
