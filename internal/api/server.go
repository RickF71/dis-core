package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"runtime"

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

	coreauth "dis-core/internal/core/authority"
	"time"

	"github.com/google/uuid"

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

	// New in MX-3.4: the core authority engine and engine config
	Engine *coreauth.Engine
	Cfg    *coreauth.Config
}

// New creates a new Server instance and wires all dependencies.
func New(db *pgxpool.Pool, led *ledger.Ledger, engine *coreauth.Engine) *Server {
	return NewWithPolicy(db, led, nil, engine)
}

// NewWithPolicy creates a new Server instance with a policy engine.
func NewWithPolicy(db *pgxpool.Pool, led *ledger.Ledger, policyEngine policy.PolicyEngine, engine *coreauth.Engine) *Server {
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

	// Wire in provided authority engine (may be nil during gradual migration)
	s.Engine = engine

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

	// Debug: print registered routes at startup to help diagnose runtime router issues
	// This duplicates the behavior of the /debug/routes handler but ensures routes
	// are visible in startup logs for live processes.
	_ = chi.Walk(s.router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		// Try to resolve a human-friendly handler name when possible.
		name := fmt.Sprintf("%T", handler)
		if hf, ok := handler.(http.HandlerFunc); ok {
			ptr := reflect.ValueOf(hf).Pointer()
			if f := runtime.FuncForPC(ptr); f != nil {
				name = f.Name()
			}
		}
		s.logger.Printf("ROUTE: %s %s -> %s", method, route, name)
		return nil
	})

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

	// Ensure DB is available (some tests run without initializing DB)
	db := s.requireDB(w)
	if db == nil {
		return
	}
	ctx := r.Context()
	var domainData string
	err := db.QueryRow(ctx, `
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

// requireDB ensures the server has an initialized DB pool. If not, it writes
// a 503 response and returns nil so callers can early-return. This is useful
// in unit tests where a DB may not be configured and avoids nil-pointer
// dereferences that would panic during pgx operations.
func (s *Server) requireDB(w http.ResponseWriter) *pgxpool.Pool {
	if s.db == nil {
		http.Error(w, "db not initialized (test mode)", http.StatusServiceUnavailable)
		return nil
	}
	return s.db
}

// handleAuthorityIntrospect serves the /api/authority/introspect endpoint
func (s *Server) handleAuthorityIntrospect(w http.ResponseWriter, r *http.Request) {
	if s.Engine == nil {
		http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
		return
	}
	ctx := r.Context()
	info, err := s.Engine.GetIntrospect(ctx)
	if err != nil {
		http.Error(w, "failed to introspect authority engine: "+err.Error(), http.StatusInternalServerError)
		return
	}
	JSON(w, http.StatusOK, info)
}

// handleAuthorityLogs is a placeholder for authority logs endpoint
func (s *Server) handleAuthorityLogs(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusNotImplemented, map[string]any{"error": "authority logs not implemented"})
}

// --- AT reflection handlers (displays AT-series state and decisions)

// handleATState returns a JSON-only representation of the current AT phase,
// available rules and last decision. This is a lightweight reflection endpoint
// and does not execute any AT logic.
func (s *Server) handleATState(w http.ResponseWriter, r *http.Request) {
	// Minimal structured response; implementations may enrich using policy engine.
	resp := map[string]any{
		"phase":  0,
		"policy": map[string]any{},
		"rules":  []any{},
		"last_decision": map[string]any{
			"allow":      true,
			"reason":     "",
			"deny_code":  "",
			"policy_ref": "",
			"details":    map[string]any{},
		},
	}
	JSON(w, http.StatusOK, resp)
}

// handleATDecisions returns recent structured decisions (deny/allow) for AT
func (s *Server) handleATDecisions(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"decisions": []any{},
	}
	JSON(w, http.StatusOK, resp)
}

// handleATRunPhase accepts a POST to request running an AT phase. This endpoint
// only acknowledges the request and returns the requested phase id; execution
// is out of scope for this reflection layer.
func (s *Server) handleATRunPhase(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		JSONError(w, http.StatusBadRequest, "phase id required")
		return
	}
	JSON(w, http.StatusOK, map[string]any{"started": true, "phase": id})
}

// handleAuthorityLineage handles /api/authority/lineage/{domain_id}
func (s *Server) handleAuthorityLineage(w http.ResponseWriter, r *http.Request) {
	if s.Engine == nil {
		http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
		return
	}
	// Try chi path param first
	domainID := chi.URLParam(r, "domain_id")
	if domainID == "" {
		// fallback to query param
		domainID = r.URL.Query().Get("id")
	}
	if domainID == "" {
		http.Error(w, "missing domain id", http.StatusBadRequest)
		return
	}
	lineage, err := s.Engine.GetLineage(r.Context(), domainID)
	if err != nil {
		http.Error(w, "failed to fetch lineage: "+err.Error(), http.StatusInternalServerError)
		return
	}
	JSON(w, http.StatusOK, lineage)
}

// handleAuthorityLineageVerify is a simple wrapper for lineage verification endpoints
func (s *Server) handleAuthorityLineageVerify(w http.ResponseWriter, r *http.Request) {
	// For now reuse the lineage response; verification will be implemented in MX-3.4
	s.handleAuthorityLineage(w, r)
}

// --- Domain freeze HTTP handlers (MX-3.9)

// handleFreezeDomain accepts JSON body {scope, reason, ttl, created_by}
// POST /api/domain/{id}/freeze
func (s *Server) handleFreezeDomain(w http.ResponseWriter, r *http.Request) {
	if s.Engine == nil {
		http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
		return
	}
	domainID := chi.URLParam(r, "id")
	if domainID == "" {
		http.Error(w, "missing domain id", http.StatusBadRequest)
		return
	}

	var req struct {
		Scope     string `json:"scope"`
		Reason    string `json:"reason"`
		TTL       string `json:"ttl"`
		CreatedBy string `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	var ttlPtr *time.Time
	if req.TTL != "" {
		t, err := time.Parse(time.RFC3339, req.TTL)
		if err != nil {
			http.Error(w, "invalid ttl (RFC3339 expected): "+err.Error(), http.StatusBadRequest)
			return
		}
		ttlPtr = &t
	}

	// Prefer authenticated actor information from request context. If the
	// ActiveUser is bound to a corporeal domain we use that UID as the
	// creator. Fall back to the optional `created_by` field in the request
	// body for tests and legacy callers.
	createdBy := uuid.Nil
	if au := auth.GetActiveUser(r); au != nil && au.IsBound() {
		if u, err := uuid.Parse(au.CorporealDomainUID); err == nil {
			createdBy = u
		}
	} else if req.CreatedBy != "" {
		if u, err := uuid.Parse(req.CreatedBy); err == nil {
			createdBy = u
		}
	}

	id, err := s.Engine.FreezeDomain(r.Context(), uuid.MustParse(domainID), req.Scope, req.Reason, createdBy, ttlPtr)
	if err != nil {
		http.Error(w, "failed to freeze domain: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSON(w, http.StatusOK, map[string]any{"freeze_id": id.String()})
}

// handleUnfreezeDomain accepts JSON body {scope, reason, created_by}
// POST /api/domain/{id}/unfreeze
func (s *Server) handleUnfreezeDomain(w http.ResponseWriter, r *http.Request) {
	if s.Engine == nil {
		http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
		return
	}
	domainID := chi.URLParam(r, "id")
	if domainID == "" {
		http.Error(w, "missing domain id", http.StatusBadRequest)
		return
	}

	var req struct {
		Scope     string `json:"scope"`
		Reason    string `json:"reason"`
		CreatedBy string `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Prefer authenticated actor; fall back to request body value.
	createdBy := uuid.Nil
	if au := auth.GetActiveUser(r); au != nil && au.IsBound() {
		if u, err := uuid.Parse(au.CorporealDomainUID); err == nil {
			createdBy = u
		}
	} else if req.CreatedBy != "" {
		if u, err := uuid.Parse(req.CreatedBy); err == nil {
			createdBy = u
		}
	}

	if err := s.Engine.UnfreezeDomain(r.Context(), uuid.MustParse(domainID), req.Scope, createdBy, req.Reason); err != nil {
		http.Error(w, "failed to unfreeze domain: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSON(w, http.StatusOK, map[string]any{"unfrozen": true})
}

// handleOverrideFreezeDomain accepts JSON body {scope, reason, prior_freeze_id, created_by}
// POST /api/domain/{id}/freeze/override
func (s *Server) handleOverrideFreezeDomain(w http.ResponseWriter, r *http.Request) {
	if s.Engine == nil {
		http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
		return
	}
	domainID := chi.URLParam(r, "id")
	if domainID == "" {
		http.Error(w, "missing domain id", http.StatusBadRequest)
		return
	}

	var req struct {
		Scope       string `json:"scope"`
		Reason      string `json:"reason"`
		PriorFreeze string `json:"prior_freeze_id"`
		CreatedBy   string `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Prefer authenticated actor; fall back to request body value.
	createdBy := uuid.Nil
	if au := auth.GetActiveUser(r); au != nil && au.IsBound() {
		if u, err := uuid.Parse(au.CorporealDomainUID); err == nil {
			createdBy = u
		}
	} else if req.CreatedBy != "" {
		if u, err := uuid.Parse(req.CreatedBy); err == nil {
			createdBy = u
		}
	}

	priorID := uuid.Nil
	if req.PriorFreeze != "" {
		if u, err := uuid.Parse(req.PriorFreeze); err == nil {
			priorID = u
		} else {
			http.Error(w, "invalid prior_freeze_id: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "prior_freeze_id required for override", http.StatusBadRequest)
		return
	}

	id, err := s.Engine.OverrideFreezeDomain(r.Context(), uuid.MustParse(domainID), req.Scope, createdBy, req.Reason, priorID)
	if err != nil {
		http.Error(w, "failed to override freeze: "+err.Error(), http.StatusInternalServerError)
		return
	}

	JSON(w, http.StatusOK, map[string]any{"freeze_id": id.String()})
}
