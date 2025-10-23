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
		{"peers", net.EnsurePeersTable},
		{"identities", db.EnsureIdentitiesSchema},
		{"handshakes", db.EnsureHandshakesSchema},
		{"schemas", schema.EnsureSchemasTable},
		{"policies", policy.EnsurePoliciesTable},
		{"overlays", overlay.EnsureOverlaysTable},
		{"import_receipts", ledger.EnsureImportReceiptsSchema},
		{"mirror_events", mirrorspin.EnsureMirrorEventsTable},
		{"receipts", db.EnsureReceiptsSchema},
	}

	for _, step := range steps {
		if err := step.fn(dbConn); err != nil {
			log.Printf("⚠️  %s table setup failed: %v", step.name, err)
			return fmt.Errorf("%s table: %w", step.name, err)
		}
		fmt.Printf("✅ %s table ready\n", step.name)
	}

	fmt.Println("✅ All tables ensured.")
	return nil
}
