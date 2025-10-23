package bootstrap

import (
	"database/sql"
	"fmt"
	"log"

	"dis-core/internal/db"
	"dis-core/internal/domain"
	"dis-core/internal/ledger"
	"dis-core/internal/mirrorspin"
	"dis-core/internal/net"
	"dis-core/internal/overlay"
	"dis-core/internal/policy"
	"dis-core/internal/schema"
)

// BootstrapAllTables ensures all core and subsystem tables exist in dependency order.
func BootstrapAllTables(dbConn *sql.DB) error {
	fmt.Println("🚀 Bootstrapping all DIS-Core tables...")

	steps := []struct {
		name string
		fn   func(*sql.DB) error
	}{
		{"domains", domain.EnsureDomainsTable},
		{"schemas", schema.EnsureSchemasTable},
		{"overlays", overlay.EnsureOverlaysTable},
		{"policies", policy.EnsurePoliciesTable},
		{"mirror_events", mirrorspin.EnsureMirrorEventsTable},
		{"peers", net.EnsurePeersTable},
		{"identities", db.EnsureIdentitiesSchema},
		{"handshakes", db.EnsureHandshakesSchema},
		{"import_receipts", ledger.EnsureImportReceiptsSchema},
		{"receipts", db.EnsureReceiptsSchema},
	}

	for _, step := range steps {
		if err := step.fn(dbConn); err != nil {
			log.Printf("⚠️  %s table setup failed: %v", step.name, err)
			return fmt.Errorf("%s table: %w", step.name, err)
		}
		log.Printf("✅ %s table ready", step.name)
	}

	fmt.Println("✅ All tables ensured.")
	return nil
}
