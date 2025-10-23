package db

import (
	"database/sql"
	"fmt"
	"log"

	"dis-core/internal/domain"
	"dis-core/internal/mirrorspin"
	"dis-core/internal/overlay"
	"dis-core/internal/policy"
	"dis-core/internal/schema"
)

// BootstrapAllTables creates all core tables in the correct order.
func BootstrapAllTables(db *sql.DB) error {
	fmt.Println("🧱 Bootstrapping database tables...")

	if err := domain.EnsureDomainsTable(db); err != nil {
		return fmt.Errorf("domains: %v", err)
	}

	if err := schema.EnsureSchemasTable(db); err != nil {
		return fmt.Errorf("schemas: %v", err)
	}

	if err := overlay.EnsureOverlaysTable(db); err != nil {
		return fmt.Errorf("overlays: %v", err)
	}

	if err := policy.EnsurePoliciesTable(db); err != nil {
		return fmt.Errorf("policies: %v", err)
	}

	if err := mirrorspin.EnsureMirrorEventsTable(db); err != nil {
		log.Printf("⚠️ MirrorSpin table setup issue: %v", err)
	}

	fmt.Println("✅ All tables ensured.")
	return nil
}
