package db

import (
	"database/sql"
	"fmt"
)

// Initialize ensures that all core DIS tables exist.
func Initialize(db *sql.DB) error {
	fmt.Println("🔧 Initializing database schema...")

	if err := EnsureIdentitiesSchema(db); err != nil {
		return fmt.Errorf("identities table: %w", err)
	}

	// Future-proof placeholders for other schemas
	// if err := EnsureHandshakesSchema(db); err != nil { ... }
	// if err := EnsureReceiptsSchema(db); err != nil { ... }

	fmt.Println("✅ Database initialization complete.")
	return nil
}
