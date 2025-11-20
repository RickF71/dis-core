package boot

import (
	"fmt"
	"log"

	"dis-core/internal/authority"
	"dis-core/internal/config"
	"dis-core/internal/identity"

	"dis-core/internal/authority/mutation"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GOV-4: Dependency injection wiring for PolicyEvaluator and AuditLogger

// MutationComponents holds the wired mutation engine components
type MutationComponents struct {
	Engine   *authority.SeatMutationEngine
	Policy   mutation.PolicyEvaluator
	Audit    mutation.AuditLogger
	EventBus authority.EventBus
}

// WireMutationEngine creates and wires the mutation engine with adapters
func WireMutationEngine(
	db *pgxpool.Pool,
	triadRepo *identity.TriadRepository,
	cfg config.PolicyAuditConfig,
) (*MutationComponents, error) {
	log.Println("🔧 Wiring GOV-4 mutation engine components...")

	// Create policy evaluator
	var policyEval mutation.PolicyEvaluator
	if cfg.PolicyEnabled {
		if cfg.PolicyLoad == "fs" {
			opa, err := mutation.NewOPAPolicyEvaluator(cfg.PolicyBundleDir, true)
			if err != nil {
				return nil, fmt.Errorf("failed to create OPA policy evaluator: %w", err)
			}
			policyEval = opa
			log.Printf("✅ Policy evaluator: OPA (bundle: %s)", cfg.PolicyBundleDir)
		} else {
			return nil, fmt.Errorf("unsupported policy load mode: %s (only 'fs' supported)", cfg.PolicyLoad)
		}
	} else {
		policyEval = &mutation.NullPolicyEvaluator{}
		log.Println("⚠️  Policy evaluator: NULL (disabled - development mode)")
	}

	// Create audit logger
	var auditLog mutation.AuditLogger
	if cfg.AuditEnabled {
		auditLog = mutation.NewDBAuditLogger(
			db,
			true,
			cfg.AuditReceiptKind,
			cfg.AuditDecisionsTable,
		)
		log.Printf("✅ Audit logger: DB (table: %s, receipt: %s)",
			cfg.AuditDecisionsTable, cfg.AuditReceiptKind)
	} else {
		auditLog = &mutation.NullAuditLogger{}
		log.Println("⚠️  Audit logger: NULL (disabled - development mode)")
	}

	// Create event bus
	eventBus := authority.NewMemoryBus()
	log.Println("✅ Event bus: In-memory (SSE)")

	// Create mutation engine
	engine := authority.NewSeatMutationEngine(
		triadRepo,
		policyEval,
		auditLog,
		eventBus,
	)
	log.Println("✅ Mutation engine: Wired and ready")

	if cfg.IsDevelopmentMode() {
		log.Println("⚠️  ⚠️  ⚠️  DEVELOPMENT MODE: Policy and audit disabled ⚠️  ⚠️  ⚠️")
	}

	return &MutationComponents{
		Engine:   engine,
		Policy:   policyEval,
		Audit:    auditLog,
		EventBus: eventBus,
	}, nil
}
