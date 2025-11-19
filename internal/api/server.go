package api

import (
	"encoding/json"
	"log"
	"net/http"

	"dis-core/internal/api/middleware"
	"dis-core/internal/auth"
	"dis-core/internal/authority"
	"dis-core/internal/corporeal"
	"dis-core/internal/identity"
	"dis-core/internal/ledger"
	"dis-core/internal/policy"
	"dis-core/internal/repo"
	"dis-core/internal/schema"
	"dis-core/internal/seats"
	"dis-core/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server is the core API instance for DIS-Core.
type Server struct {
	db      *pgxpool.Pool
	ledger  *ledger.Ledger
	router  *chi.Mux
	schemas *schema.Registry
	logger  *log.Logger
	policy  policy.PolicyEngine

	authoritySchema  *authority.AuthoritySchema
	authorityConsole *authority.Console

	// Phase S: Seats management
	seatsRepo    *seats.Repository
	seatsService *seats.Service

	// GOV-1: Identity triad management
	triadRepo *identity.TriadRepository

	// GOV-3: Seat mutations, audits, events
	mutationEngine *authority.SeatMutationEngine
	eventBus       authority.EventBus

	// GOV-11G: Identity receipt recording for schema adoption and policy updates
	identityReceiptRecorder *identity.IdentityReceiptStore

	// GOV-13: Contract service for DSCI contract management
	contractService services.ContractService

	// Phase 0-R.5: Corporeal + Actor-Domain Bootstrap
	corporealBootstrapper *corporeal.Bootstrapper

	// Auth Challenge Store (Phase: Auth Challenge Refactor)
	challengeStore auth.ChallengeStore
}

// New creates a new Server instance and wires all dependencies.
func New(db *pgxpool.Pool, led *ledger.Ledger) *Server {
	return NewWithPolicy(db, led, nil)
}

// NewWithPolicy creates a new Server instance with a policy engine.
func NewWithPolicy(db *pgxpool.Pool, led *ledger.Ledger, policyEngine policy.PolicyEngine) *Server {
	s := &Server{
		db:             db,
		ledger:         led,
		router:         NewFormatAwareRouter(),
		logger:         log.Default(),
		policy:         policyEngine,
		challengeStore: auth.NewPostgresChallengeStore(db),
	}

	authzSchema, err := authority.LoadSchema("./internal/schema/schema.json")
	if err != nil {
		s.logger.Printf("failed to load authority schema: %v", err)
	}

	s.authoritySchema = authzSchema
	s.authorityConsole = authority.NewConsole(db, led, nil, s.authoritySchema)

	// Phase S: Initialize seats repository and service
	s.seatsRepo = seats.NewRepository(db)
	s.seatsService = seats.NewService(s.seatsRepo)

	// GOV-1: Initialize identity triad repository
	s.triadRepo = identity.NewTriadRepository(db)

	// GOV-3: Initialize event bus and mutation engine
	s.eventBus = authority.NewMemoryBus()
	// Note: mutationEngine requires PolicyEvaluator and AuditLogger which are not yet wired
	// Will be initialized separately after policy engine is ready
	s.mutationEngine = nil // Set via SetMutationEngine() after boot

	// GOV-11G: Initialize identity receipt recorder
	s.identityReceiptRecorder = identity.NewIdentityReceiptStore(db)

	// GOV-13: Initialize contract service
	contractRepo := repo.NewContractRepository(db)
	s.contractService = services.NewContractService(contractRepo, s.identityReceiptRecorder)

	// Phase 0-R.5: Initialize corporeal bootstrapper
	s.corporealBootstrapper = corporeal.NewBootstrapper(db)

	// Phase 6: Attach universal middleware stack before registering routes
	middleware.Attach(s.router, s.db, s.policy)

	s.RegisterAllRoutes()

	// Install format consistency checker
	s.InstallFormatConsistencyCheck()

	return s
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "running",
		"service":   "dis-core",
		"ledger":    s.ledger != nil,
		"authority": s.authorityConsole != nil,
		"schemas":   s.schemas != nil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleDomainGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	domainID := r.URL.Path[len("/api/domain/"):]
	if domainID == "" {
		http.Error(w, "domain ID required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	var domainData string
	err := s.db.QueryRow(ctx, `
SELECT content FROM canon
WHERE type = 'domain' AND content->>'domain_id' = $1
LIMIT 1
`, domainID).Scan(&domainData)

	if err != nil {
		s.logger.Printf("domain query failed: %v", err)
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(domainData))
}

// SetMutationEngine sets the mutation engine for GOV-3 seat transitions
func (s *Server) SetMutationEngine(engine *authority.SeatMutationEngine) {
	s.mutationEngine = engine
}

func (s *Server) Handler() *chi.Mux      { return s.router }
func (s *Server) DB() *pgxpool.Pool      { return s.db }
func (s *Server) Ledger() *ledger.Ledger { return s.ledger }
func (s *Server) Router() *chi.Mux       { return s.router }
