package policy

import (
	"context"
	"database/sql"
	"fmt"
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
