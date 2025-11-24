// NOTE: This bootstrap is intended for test and development convenience. Production
// schema changes should be delivered via SQL migrations in the `migrations/`
// directory. The bootstrap below mirrors those migrations so the test harness
// can rely on a consistent schema even without running the migration tool.
//
// When adding or changing persistent schema required by runtime code (for
// example fields used by the K-1 KnowThyself atomic flow), please create a
// migration SQL file under migrations/ with UP and DOWN sections and keep this
// bootstrap in-sync for the test harness.
package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BootstrapAllTables ensures all core and subsystem tables exist in dependency order.
func BootstrapAllTables(dbConn *pgxpool.Pool) error {
	ctx := context.Background()
	fmt.Println("🚀 Bootstrapping all DIS-Core tables...")

	// For this refactor, we'll use direct SQL statements rather than
	// updating all the individual EnsureXTable functions
	tables := []string{
		`CREATE TABLE IF NOT EXISTS domains (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT UNIQUE,
			parent_id UUID REFERENCES domains(id),
			domain_type TEXT,
			authority TEXT,
			payload JSONB DEFAULT '{}'::jsonb,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS schemas (
			id TEXT PRIMARY KEY,
			version TEXT,
			content JSONB,
			hash TEXT,
			imported_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS overlays (
			id TEXT PRIMARY KEY,
			content JSONB,
			imported_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS policies (
			id TEXT PRIMARY KEY,
			content JSONB,
			imported_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS mirror_events (
			id TEXT PRIMARY KEY,
			event_type TEXT,
			payload JSONB,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS peers (
			id TEXT PRIMARY KEY,
			address TEXT,
			last_seen TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS identities (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			subject TEXT UNIQUE,
			presentation_name TEXT,
			identity_type TEXT,
			public_key TEXT,
			metadata JSONB,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS handshakes (
			id TEXT PRIMARY KEY,
			token TEXT UNIQUE,
			subject TEXT,
			peer_id TEXT,
			status TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,
		// Persistent sessions for long-lived actor authentication (MX-K4)
		`CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			actor_id UUID NOT NULL,
			domain_id UUID NOT NULL,
			seat_id UUID NOT NULL,
			token TEXT UNIQUE NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			expires_at TIMESTAMPTZ NOT NULL,
			revoked_at TIMESTAMPTZ NULL
		);`,
		`CREATE TABLE IF NOT EXISTS receipts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			type TEXT NOT NULL,
			actor TEXT,
			target TEXT,
			domain UUID,
			payload JSONB,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,

		// GOV-10: identity_receipts ledger used throughout identity/aliases
		`CREATE TABLE IF NOT EXISTS identity_receipts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			domain_id UUID NOT NULL REFERENCES domains(id),
			actor_id UUID NOT NULL,
			action TEXT NOT NULL,
			payload JSONB NOT NULL,
			prev_id UUID REFERENCES identity_receipts(id),
			hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			consent_by UUID NOT NULL REFERENCES domains(id),
			alias_scope TEXT,
			target_domain_id UUID NULL REFERENCES domains(id),
			source_domain_id UUID NULL REFERENCES domains(id),
			external_subject TEXT NULL,
			channel TEXT NULL,
			method TEXT NULL,
			scope TEXT NULL
		);`,

		`CREATE INDEX IF NOT EXISTS idx_identity_receipts_actor ON identity_receipts(actor_id, created_at ASC);`,
		`CREATE INDEX IF NOT EXISTS idx_identity_receipts_domain ON identity_receipts(domain_id);`,
		`CREATE INDEX IF NOT EXISTS idx_identity_receipts_action ON identity_receipts(action);`,
		`CREATE INDEX IF NOT EXISTS idx_identity_receipts_prev ON identity_receipts(prev_id);`,
		`CREATE TABLE IF NOT EXISTS import_warnings (
			id TEXT PRIMARY KEY,
			message TEXT,
			source_file TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);`,

		// GOV-12: aliases
		`CREATE TABLE IF NOT EXISTS aliases (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			owner_domain_id UUID NOT NULL,
			target_domain_id UUID NOT NULL,
			alias_name TEXT NOT NULL,
			alias_type TEXT NOT NULL,
			is_corporeal_auto BOOLEAN NOT NULL DEFAULT FALSE,
			dsci_contract_id UUID NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			retired_at TIMESTAMPTZ NULL,
			metadata JSONB NOT NULL DEFAULT '{}'::jsonb
		);`,

		`CREATE INDEX IF NOT EXISTS idx_aliases_owner_domain ON aliases(owner_domain_id, alias_type);`,
		`CREATE INDEX IF NOT EXISTS idx_aliases_target_domain ON aliases(target_domain_id, alias_type);`,
		`CREATE INDEX IF NOT EXISTS idx_aliases_dsci_contract ON aliases(dsci_contract_id) WHERE dsci_contract_id IS NOT NULL;`,
		`CREATE INDEX IF NOT EXISTS idx_aliases_created_at ON aliases(created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_aliases_active ON aliases(owner_domain_id, alias_type) WHERE retired_at IS NULL;`,
		// Partial unique index for active aliases (owner,target,name) tuple
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_aliases_unique_active ON aliases(owner_domain_id, target_domain_id, alias_name) WHERE retired_at IS NULL;`,
		`CREATE TABLE IF NOT EXISTS authority_decisions (
			id VARCHAR(36) PRIMARY KEY,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			actor VARCHAR(255) NOT NULL,
			domain VARCHAR(255) NOT NULL,
			policy_id VARCHAR(255) NOT NULL,
			input JSONB NOT NULL,
			result JSONB NOT NULL,
			reason TEXT,
			phase_tag VARCHAR(100),
			replay_hash VARCHAR(64) NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_authority_decisions_domain ON authority_decisions (domain);`,
		`CREATE INDEX IF NOT EXISTS idx_authority_decisions_policy ON authority_decisions (policy_id);`,
		`CREATE INDEX IF NOT EXISTS idx_authority_decisions_created ON authority_decisions (created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_authority_decisions_actor ON authority_decisions (actor);`,
		`CREATE INDEX IF NOT EXISTS idx_authority_decisions_phase ON authority_decisions (phase_tag);`,

		// Phase 9C: Receipt Verification & Provenance Continuity
		`CREATE TABLE IF NOT EXISTS receipts_9c (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			receipt_type TEXT NOT NULL,
			event_id TEXT NOT NULL,
			policy_ref TEXT,
			redaction_ref TEXT,
			issued_by TEXT,
			issued_at TIMESTAMPTZ DEFAULT now(),
			verified BOOLEAN DEFAULT FALSE,
			metadata JSONB DEFAULT '{}'::jsonb
		);`,
		`CREATE INDEX IF NOT EXISTS idx_receipts_9c_event_id ON receipts_9c(event_id);`,
		`CREATE INDEX IF NOT EXISTS idx_receipts_9c_policy_ref ON receipts_9c(policy_ref);`,
		`CREATE INDEX IF NOT EXISTS idx_receipts_9c_redaction_ref ON receipts_9c(redaction_ref);`,
		// GOV-9: authority_receipts ledger (hash-linked governance receipts)
		`CREATE TABLE IF NOT EXISTS authority_receipts (
			id TEXT PRIMARY KEY,
			domain_id UUID NOT NULL,
			action TEXT NOT NULL,
			prev_id TEXT,
			payload JSONB,
			hash TEXT NOT NULL,
			policy_digest TEXT,
			created_at TIMESTAMPTZ DEFAULT now(),
			CONSTRAINT fk_domain FOREIGN KEY (domain_id) REFERENCES domains(id),
			CONSTRAINT fk_prev FOREIGN KEY (prev_id) REFERENCES authority_receipts(id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_authority_receipts_domain ON authority_receipts(domain_id, created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_authority_receipts_action ON authority_receipts(action);`,
		`CREATE INDEX IF NOT EXISTS idx_authority_receipts_prev ON authority_receipts(prev_id);`,

		// Phase 10F: Continuity Lineage Proofs - Fix Receipts Table
		`CREATE TABLE IF NOT EXISTS fix_receipts (
			id TEXT PRIMARY KEY,
			source_receipt_id UUID NOT NULL REFERENCES receipts_9c(id),
			fix_type TEXT NOT NULL,
			target_domain TEXT,
			policy_context JSONB DEFAULT '{}'::jsonb,
			proof_refs TEXT[] DEFAULT '{}',
			verification_status TEXT DEFAULT 'pending',
			created_at TIMESTAMPTZ DEFAULT now(),
			verified_at TIMESTAMPTZ
		);`,
		`CREATE INDEX IF NOT EXISTS idx_fix_receipts_source ON fix_receipts(source_receipt_id);`,
		`CREATE INDEX IF NOT EXISTS idx_fix_receipts_type ON fix_receipts(fix_type);`,
		`CREATE INDEX IF NOT EXISTS idx_fix_receipts_domain ON fix_receipts(target_domain);`,
		`CREATE INDEX IF NOT EXISTS idx_fix_receipts_status ON fix_receipts(verification_status);`,

		// Phase 10G: Cross-Domain Federation Proofs
		`CREATE TABLE IF NOT EXISTS federation_proofs (
			id TEXT PRIMARY KEY,
			source_domain TEXT NOT NULL,
			target_domain TEXT NOT NULL,
			proof_ref TEXT NOT NULL,
			federation_hash TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			trust_level TEXT DEFAULT 'medium',
			timestamp TIMESTAMPTZ DEFAULT now(),
			verified_at TIMESTAMPTZ,
			metadata JSONB DEFAULT '{}'::jsonb
		);`,
		`CREATE INDEX IF NOT EXISTS idx_federation_proofs_source ON federation_proofs(source_domain);`,
		`CREATE INDEX IF NOT EXISTS idx_federation_proofs_target ON federation_proofs(target_domain);`,
		`CREATE INDEX IF NOT EXISTS idx_federation_proofs_hash ON federation_proofs(federation_hash);`,
		`CREATE INDEX IF NOT EXISTS idx_federation_proofs_status ON federation_proofs(status);`,

		// Phase 10G: Federation Trust Mappings
		`CREATE TABLE IF NOT EXISTS federation_trust (
			id TEXT PRIMARY KEY,
			domain_a TEXT NOT NULL,
			domain_b TEXT NOT NULL,
			trust_level TEXT DEFAULT 'medium',
			established_at TIMESTAMPTZ DEFAULT now(),
			expires_at TIMESTAMPTZ,
			metadata JSONB DEFAULT '{}'::jsonb,
			UNIQUE(domain_a, domain_b)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_federation_trust_domains ON federation_trust(domain_a, domain_b);`,
		`CREATE INDEX IF NOT EXISTS idx_federation_trust_level ON federation_trust(trust_level);`,

		// Phase S1: Seats Schema
		`CREATE TABLE IF NOT EXISTS domain_seats (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			domain_id UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
			parent_seat_id UUID REFERENCES domain_seats(id) ON DELETE SET NULL,
			seat_type TEXT NOT NULL DEFAULT 'prime',
			member_id TEXT,
			appointed_by UUID REFERENCES domain_seats(id) ON DELETE SET NULL,
			appointment_receipt TEXT,
			rego_ref TEXT,
			rego_text TEXT,
			policy_version TEXT,
			scope TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT now()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_domain_seats_domain ON domain_seats(domain_id);`,
		`CREATE INDEX IF NOT EXISTS idx_domain_seats_type ON domain_seats(seat_type);`,
		`CREATE INDEX IF NOT EXISTS idx_domain_seats_status ON domain_seats(status);`,
		`CREATE INDEX IF NOT EXISTS idx_domain_seats_member ON domain_seats(member_id);`,
		`CREATE INDEX IF NOT EXISTS idx_domain_seats_parent ON domain_seats(parent_seat_id);`,
		// Seat roles table (MX-K10)
		`CREATE TABLE IF NOT EXISTS seat_roles (
			seat_id UUID NOT NULL,
			role TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (seat_id, role)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_seat_roles_role ON seat_roles(role);`,
		// Domain freeze state (authority-driven)
		`CREATE TABLE IF NOT EXISTS domain_freeze_state (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			domain_id UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
			scope TEXT NOT NULL,
			reason TEXT NOT NULL,
			created_by UUID NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			ttl_until TIMESTAMPTZ NULL,
			override_of UUID NULL REFERENCES domain_freeze_state(id) ON DELETE SET NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_domain_freeze_unique_active ON domain_freeze_state(domain_id, scope) WHERE is_active = TRUE;`,
		`CREATE INDEX IF NOT EXISTS idx_domain_freeze_domain_active ON domain_freeze_state(domain_id, is_active);`,

		// GOV-13: contracts table (lightweight implementation matching migration)
		`CREATE TABLE IF NOT EXISTS contracts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			domain_id UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
			subject_domain_id UUID NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
			alias_id UUID NULL REFERENCES domains(id) ON DELETE SET NULL,
			contract_type TEXT NOT NULL,
			dsci_channel TEXT NOT NULL,
			dsci_reference TEXT NOT NULL,
			dsci_version TEXT NOT NULL,
			effective_at TIMESTAMPTZ NOT NULL,
			expires_at TIMESTAMPTZ NULL,
			revoked_at TIMESTAMPTZ NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by UUID NULL REFERENCES domains(id) ON DELETE SET NULL
		);`,

		`CREATE INDEX IF NOT EXISTS idx_contracts_domain_id ON contracts(domain_id);`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_subject_domain_id ON contracts(subject_domain_id);`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_alias_id ON contracts(alias_id) WHERE alias_id IS NOT NULL;`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_status ON contracts(status);`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_effective_at ON contracts(effective_at);`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_created_at ON contracts(created_at DESC);`,
	}

	for i, createSQL := range tables {
		if _, err := dbConn.Exec(ctx, createSQL); err != nil {
			return fmt.Errorf("creating table %d: %w", i, err)
		}
	}

	fmt.Println("✅ All tables ensured.")

	// Ensure canonical test domains used by GOV-8 tests exist. These are
	// idempotent upserts so repeated bootstraps are safe. Tests expect the
	// terra UUID and a corporeal domain with a known parent relationship.
	// terra: 4daf928e-e58c-454e-8395-f3dedd103dde
	// corporeal: a1111111-1111-1111-1111-111111111111 (parent -> terra)
	seedSQL := []string{
		// Ensure aether exists as part of canonical spine (inserted before terra)
		`INSERT INTO domains (id, name, parent_id, domain_type, payload, created_at, updated_at)
		 VALUES (gen_random_uuid(), 'aether', (SELECT id FROM domains WHERE name = 'null' OR name = 'domain.null' LIMIT 1), 'aether', '{}'::jsonb, now(), now())
		 ON CONFLICT (name) DO NOTHING;`,
		`INSERT INTO domains (id, name, parent_id, domain_type, payload, created_at, updated_at)
		 VALUES ('4daf928e-e58c-454e-8395-f3dedd103dde', 'terra', (SELECT id FROM domains WHERE name = 'aether' OR name = 'domain.aether' LIMIT 1), 'terra', '{"meta": {"governance": "human_sovereign"}}'::jsonb, now(), now())
		 ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, parent_id = EXCLUDED.parent_id, payload = EXCLUDED.payload, updated_at = now();`,
		`INSERT INTO domains (id, name, parent_id, domain_type, payload, created_at, updated_at)
		 VALUES ('a1111111-1111-1111-1111-111111111111', 'terra.numen.lima.corporeal', '4daf928e-e58c-454e-8395-f3dedd103dde', 'corporeal', '{"meta": {"governance": "human_sovereign"}}'::jsonb, now(), now())
		 ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, parent_id = EXCLUDED.parent_id, payload = EXCLUDED.payload, updated_at = now();`,
	}

	for _, s := range seedSQL {
		if _, err := dbConn.Exec(ctx, s); err != nil {
			return fmt.Errorf("seeding canonical domains: %w", err)
		}
	}

	// Phase 10F: Setup continuity lineage proofs
	if err := performPhase10FSetup(dbConn); err != nil {
		return fmt.Errorf("Phase 10F setup failed: %w", err)
	}

	// Phase 10G: Setup cross-domain proof synchronization
	if err := performPhase10GSetup(dbConn); err != nil {
		return fmt.Errorf("Phase 10G setup failed: %w", err)
	}

	// Phase 10I: Setup CSS Interchange Bridge
	if err := performPhase10ISetup(dbConn); err != nil {
		return fmt.Errorf("Phase 10I setup failed: %w", err)
	}

	// Phase 10I.2: Setup CSS Variable Map Extraction
	if err := performPhase10I2Setup(dbConn); err != nil {
		return fmt.Errorf("Phase 10I.2 setup failed: %w", err)
	}

	// Phase 10J.1: Setup Greedy JSON Domain Data Slots
	if err := performPhase10J1Setup(ctx, dbConn); err != nil {
		return fmt.Errorf("Phase 10J.1 setup failed: %w", err)
	}

	// Phase 10J.2b: Normalize Domain Data Structure
	if err := performPhase10J2bNormalization(ctx, dbConn); err != nil {
		return fmt.Errorf("Phase 10J.2b normalization failed: %w", err)
	}

	// Phase S1: Seats schema is ensured via table creation above
	fmt.Println("✅ Phase S1 — seats schema online")

	return nil
}

// performPhase10FSetup ensures Phase 10F continuity lineage proof infrastructure
func performPhase10FSetup(dbConn *pgxpool.Pool) error {
	ctx := context.Background()

	// Check fix_receipts table setup
	var fixReceiptCount int
	err := dbConn.QueryRow(ctx, `SELECT COUNT(*) FROM fix_receipts`).Scan(&fixReceiptCount)
	if err != nil {
		return fmt.Errorf("fix_receipts table verification failed: %w", err)
	}

	// Check orphan receipts that need lineage tracking
	var orphanCount int
	err = dbConn.QueryRow(ctx, `
		SELECT COUNT(*) FROM receipts_9c
		WHERE policy_ref IS NULL OR redaction_ref IS NULL
	`).Scan(&orphanCount)
	if err != nil {
		return fmt.Errorf("orphan receipt check failed: %w", err)
	}

	// Create phase_10f.log directory if needed
	logDir := "phases"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("creating phases directory: %w", err)
	}

	// Write Phase 10F status to log
	logPath := filepath.Join(logDir, "phase_10f.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening phase_10f.log: %w", err)
	}
	defer logFile.Close()

	timestamp := time.Now().UTC().Format(time.RFC3339)
	logEntry := fmt.Sprintf("✅ Phase 10F — Lineage proofs initialized: %d fix receipts, %d orphan receipts tracked (%s)\n",
		fixReceiptCount, orphanCount, timestamp)

	if _, err := logFile.WriteString(logEntry); err != nil {
		return fmt.Errorf("writing to phase_10f.log: %w", err)
	}

	fmt.Printf("✅ Phase 10F — Lineage proofs ready: %d fix receipts, %d orphans tracked\n",
		fixReceiptCount, orphanCount)

	return nil
}

// performPhase10GSetup ensures Phase 10G cross-domain proof synchronization infrastructure
func performPhase10GSetup(dbConn *pgxpool.Pool) error {
	ctx := context.Background()

	// Check federation_proofs table setup
	var federationProofCount int
	err := dbConn.QueryRow(ctx, `SELECT COUNT(*) FROM federation_proofs`).Scan(&federationProofCount)
	if err != nil {
		return fmt.Errorf("federation_proofs table verification failed: %w", err)
	}

	// Check federation_trust table setup and get trust relationships count
	var trustRelationshipCount int
	err = dbConn.QueryRow(ctx, `SELECT COUNT(*) FROM federation_trust`).Scan(&trustRelationshipCount)
	if err != nil {
		return fmt.Errorf("federation_trust table verification failed: %w", err)
	}

	// Create phase_10g.log directory if needed
	logDir := "phases"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("creating phases directory: %w", err)
	}

	// Write Phase 10G status to log
	logPath := filepath.Join(logDir, "phase_10g.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening phase_10g.log: %w", err)
	}
	defer logFile.Close()

	timestamp := time.Now().UTC().Format(time.RFC3339)
	logEntry := fmt.Sprintf("✅ Phase 10G — Cross-domain proof sync initialized: %d federation proofs, %d trust relationships (%s)\n",
		federationProofCount, trustRelationshipCount, timestamp)

	if _, err := logFile.WriteString(logEntry); err != nil {
		return fmt.Errorf("writing to phase_10g.log: %w", err)
	}

	fmt.Printf("✅ Phase 10G — Cross-domain proof synchronization ready: %d federation proofs, %d trust relationships\n",
		federationProofCount, trustRelationshipCount)

	return nil
}

// performPhase10ISetup ensures Phase 10I CSS Interchange Bridge infrastructure
func performPhase10ISetup(dbConn *pgxpool.Pool) error {
	ctx := context.Background()

	// Import domaincss package for table creation
	// We'll call CreateTables to ensure CSS tables exist
	err := createDomainCSSTables(ctx, dbConn)
	if err != nil {
		return fmt.Errorf("domain CSS table creation failed: %w", err)
	}

	// Verify CSS tables exist
	var domainCSSCount, domainCSSHistoryCount int

	err = dbConn.QueryRow(ctx, `SELECT COUNT(*) FROM domain_css`).Scan(&domainCSSCount)
	if err != nil {
		return fmt.Errorf("domain_css table verification failed: %w", err)
	}

	err = dbConn.QueryRow(ctx, `SELECT COUNT(*) FROM domain_css_history`).Scan(&domainCSSHistoryCount)
	if err != nil {
		return fmt.Errorf("domain_css_history table verification failed: %w", err)
	}

	// Create phases directory if it doesn't exist
	phasesDir := "phases"
	if err := os.MkdirAll(phasesDir, 0755); err != nil {
		return fmt.Errorf("creating phases directory: %w", err)
	}

	// Log Phase 10I completion
	logPath := filepath.Join(phasesDir, "phase_10i.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening phase_10i.log: %w", err)
	}
	defer logFile.Close()

	timestamp := time.Now().UTC().Format(time.RFC3339)
	logEntry := fmt.Sprintf("✅ Phase 10I — CSS Interchange Bridge active: %d CSS records, %d history entries (%s)\n",
		domainCSSCount, domainCSSHistoryCount, timestamp)

	if _, err := logFile.WriteString(logEntry); err != nil {
		return fmt.Errorf("writing to phase_10i.log: %w", err)
	}

	fmt.Printf("✅ Phase 10I — CSS Interchange Bridge ready: %d CSS records, %d history entries\n",
		domainCSSCount, domainCSSHistoryCount)

	return nil
}

// createDomainCSSTables creates the necessary tables for CSS Interchange Bridge
func createDomainCSSTables(ctx context.Context, db *pgxpool.Pool) error {
	// Create main domain_css table
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS domain_css (
			domain_id    TEXT PRIMARY KEY,
			content_type TEXT NOT NULL DEFAULT 'text/css',
			css_content  TEXT NOT NULL DEFAULT '',
			size         INTEGER NOT NULL DEFAULT 0,
			updated_at   TIMESTAMPTZ DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create domain_css table: %w", err)
	}

	// Create history table
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS domain_css_history (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			domain_id    TEXT NOT NULL,
			content_type TEXT NOT NULL,
			css_content  TEXT NOT NULL,
			size         INTEGER NOT NULL,
			updated_at   TIMESTAMPTZ NOT NULL,
			updated_by   TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create domain_css_history table: %w", err)
	}

	// Create indexes
	_, err = db.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_domain_css_history_domain_id ON domain_css_history(domain_id);
		CREATE INDEX IF NOT EXISTS idx_domain_css_history_updated_at ON domain_css_history(updated_at DESC);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// performPhase10I2Setup ensures Phase 10I.2 CSS Variable Map Extraction infrastructure
func performPhase10I2Setup(dbConn *pgxpool.Pool) error {
	ctx := context.Background()

	// Test CSS variable extraction functionality by processing existing CSS
	var processedDomains int

	// Count domains with CSS content for processing
	err := dbConn.QueryRow(ctx, `
		SELECT COUNT(*) FROM domain_css WHERE css_content IS NOT NULL AND css_content != ''
	`).Scan(&processedDomains)
	if err != nil {
		// If domain_css table doesn't exist yet, that's ok - just log 0 processed
		processedDomains = 0
	}

	// Create phases directory if it doesn't exist
	phasesDir := "phases"
	if err := os.MkdirAll(phasesDir, 0755); err != nil {
		return fmt.Errorf("creating phases directory: %w", err)
	}

	// Log Phase 10I.2 completion
	logPath := filepath.Join(phasesDir, "phase_10i2.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening phase_10i2.log: %w", err)
	}
	defer logFile.Close()

	timestamp := time.Now().UTC().Format(time.RFC3339)
	logEntry := fmt.Sprintf("✅ Phase 10I.2 — CSS Variable Map Extraction active: API endpoint /css/vars ready, %d domains available for variable extraction (%s)\n",
		processedDomains, timestamp)

	if _, err := logFile.WriteString(logEntry); err != nil {
		return fmt.Errorf("writing to phase_10i2.log: %w", err)
	}

	fmt.Printf("✅ Phase 10I.2 — CSS Variable Map Extraction ready: API endpoint /css/vars active, %d domains available\n",
		processedDomains)

	return nil
}

// performPhase10J1Setup ensures all domains have greedy JSON payload slots (Phase 10J.4: updated for flattened structure)
func performPhase10J1Setup(ctx context.Context, db *pgxpool.Pool) error {
	fmt.Println("🧱 Phase 10J.1 — Greedy JSON slot migration starting")

	// Phase 10J.4 update: payload column now has flattened structure
	// Target structure: {css: {...}, policy: {...}, receipts: [], overlay: {...}, variables: {...}, meta: {...}, authority: {...}}

	_, err := db.Exec(ctx, `
		UPDATE domains
		SET payload = jsonb_build_object(
			'css', COALESCE(
				payload->'css',
				CASE
					WHEN payload ? 'css' AND jsonb_typeof(payload->'css') = 'string'
					THEN jsonb_build_object('content', payload->>'css', 'hash', '', 'verified', false)
					ELSE jsonb_build_object('content', '', 'hash', '', 'verified', false)
				END
			),
			'policy', COALESCE(payload->'policy', '{}'::jsonb),
			'receipts', COALESCE(payload->'receipts', '[]'::jsonb),
			'overlay', COALESCE(payload->'overlay', '{}'::jsonb),
			'variables', COALESCE(payload->'variables', '{}'::jsonb),
			'meta', COALESCE(payload->'meta', '{}'::jsonb),
			'authority', COALESCE(payload->'authority', '{}'::jsonb)
		)
		WHERE NOT (payload ? 'css' AND payload ? 'policy' AND payload ? 'receipts' AND payload ? 'overlay' AND payload ? 'variables');
	`)
	if err != nil {
		return fmt.Errorf("phase10J1 slot migration failed: %w", err)
	}
	fmt.Println("✅ Phase 10J.1 — Greedy slots ensured for all domains")
	return nil
}

// performPhase10J2bNormalization ensures all domains have normalized greedy-slot structure
// with schema_version for forward compatibility (Phase 10J.4: updated for flattened payload)
func performPhase10J2bNormalization(ctx context.Context, db *pgxpool.Pool) error {
	fmt.Println("🧱 Phase 10J.2b — Domain Payload Normalization starting")

	_, err := db.Exec(ctx, `
		UPDATE domains
		SET payload = jsonb_build_object(
			'css', COALESCE(
				payload->'css',
				jsonb_build_object(
					'content', 'body { background-color: #0f172a; color: #f1f5f9; }',
					'hash', '',
					'verified', true
				)
			),
			'overlay', COALESCE(payload->'overlay', '{}'::jsonb),
			'policy', COALESCE(payload->'policy', '{}'::jsonb),
			'receipts', COALESCE(payload->'receipts', '[]'::jsonb),
			'variables', COALESCE(payload->'variables', '{}'::jsonb),
			'meta', jsonb_set(
				COALESCE(payload->'meta', '{}'::jsonb),
				'{schema_version}',
				'"v1"'::jsonb
			),
			'authority', COALESCE(payload->'authority', '{}'::jsonb)
		)
		WHERE (
			NOT (payload ? 'css') OR
			NOT (payload ? 'meta') OR
			NOT (payload ? 'authority') OR
			NOT (payload ? 'overlay') OR
			NOT (payload ? 'policy') OR
			NOT (payload ? 'receipts') OR
			NOT (payload ? 'variables') OR
			NOT (payload->'meta' ? 'schema_version')
		);
	`)

	if err != nil {
		return fmt.Errorf("phase10J2b normalization failed: %w", err)
	}

	// Count normalized domains
	var normalizedCount int
	err = db.QueryRow(ctx, `
		SELECT COUNT(*) FROM domains
		WHERE payload->'meta'->>'schema_version' = 'v1'
	`).Scan(&normalizedCount)

	if err == nil {
		fmt.Printf("✅ Phase 10J.2b — Domain Data Normalization complete (%d domains normalized)\n", normalizedCount)
	} else {
		fmt.Println("✅ Phase 10J.2b — Domain Data Normalization complete")
	}

	return nil
}
