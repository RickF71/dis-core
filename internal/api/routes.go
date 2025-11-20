package api

import (
	"net/http"

	authorityapi "dis-core/internal/api/authority"
	"dis-core/internal/api/handlers"
	"dis-core/internal/auth"
)

// RegisterAllRoutes wires all endpoint groups into the server router with format-aware support.
// Phase 5 implementation using Chi router exclusively.
func (s *Server) RegisterAllRoutes() {
	r := s.router

	// Phase 10I: Apply CSS validation middleware to all routes (MUST be before routes)
	r.Use(CSSValidationMiddleware)

	// Phase K-1: Know Thyself atomic invite accept endpoint
	// Place invite accept AFTER middleware registration to avoid Chi panic
	r.Post("/api/invite/accept", s.handleInviteAccept)

	// External Authentication endpoints (sovereign identity)
	r.Get("/api/whoami", auth.HandleWhoAmI)
	r.Get("/api/whoami/external", auth.HandleWhoAmIExternal) // DEV ONLY
	r.Get("/api/me", s.handleMe)                             // Sprint Step 1: Read-only identity surface
	r.Get("/api/me/actors", s.handleMeActors)                // List all actors/seats for authenticated user
	r.Post("/api/me/active-actor", s.handleSetActiveActor)   // Set active actor/seat
	r.Get("/api/me/active-actor", s.handleGetActiveActor)    // Get current active actor/seat

	// Dev Auth endpoints (None Space identity selection)
	// Wrap dev handlers so they fetch the DB at request time and return 503 when no DB is configured
	r.Get("/api/auth/dev-users", func(w http.ResponseWriter, r *http.Request) {
		db := s.requireDB(w)
		if db == nil {
			return
		}
		auth.HandleDevUsers(db)(w, r)
	}) // DEV ONLY: List available dev users

	r.Post("/api/auth/dev-login", func(w http.ResponseWriter, r *http.Request) {
		db := s.requireDB(w)
		if db == nil {
			return
		}
		auth.HandleDevLogin(db)(w, r)
	}) // DEV ONLY: Validate dev login

	// Auth challenge endpoints (new canonical handlers)
	r.Post("/api/auth/challenge", auth.NewChallengeCreateHandler(s.challengeStore).ServeHTTP)
	r.Get("/api/auth/challenge/{id}/status", auth.NewChallengeStatusHandler(s.challengeStore).ServeHTTP)

	// Phase 5.5 requirement: Use WithFormatGuard for format-restricted endpoints
	// Domain CSS with format guard (File, Text, JSON only)
	allowedFormats := []Format{FormatFile, FormatText, FormatJSON}
	r.Get("/api/domain/{id}/css",
		WithFormatGuard(s.handleDomainCSSFormatAware, allowedFormats))

	// Register format-specific routes for CSS endpoint
	for _, format := range allowedFormats {
		formatPattern := "/api/domain/{id}/css/" + string(format)
		r.Get(formatPattern,
			WithFormatGuard(s.handleDomainCSSFormatAware, allowedFormats))
	}

	// Phase 10I: Register CSS Interchange Bridge routes
	s.registerDomainCSSBridgeRoutes()

	// Also register unsupported formats to test the guard
	unsupportedFormats := []Format{FormatCBOR, FormatEncrypted}
	for _, format := range unsupportedFormats {
		formatPattern := "/api/domain/{id}/css/" + string(format)
		r.Get(formatPattern,
			WithFormatGuard(s.handleDomainCSSFormatAware, allowedFormats))
	}

	// Phase 5 requirement: Core routes with format support (including CBOR and Encrypted for proper error handling)
	s.RegisterFormatAwareRoute(r, http.MethodGet, "/api/ping", s.handlePingChi, []Format{FormatJSON, FormatText, FormatFile, FormatCBOR, FormatEncrypted})
	s.RegisterFormatAwareRoute(r, http.MethodGet, "/api/health", s.handleHealthChi, []Format{FormatJSON, FormatText, FormatCBOR, FormatEncrypted})
	s.RegisterFormatAwareRoute(r, http.MethodGet, "/api/status", s.handleStatusChi, []Format{FormatJSON, FormatText, FormatCBOR, FormatEncrypted})

	// Additional format-aware routes
	s.RegisterFormatAwareRoute(r, http.MethodGet, "/api/authority/status", s.handleAuthorityStatusChi, []Format{FormatJSON, FormatText})

	// Phase 7 Authority Console routes
	s.RegisterFormatAwareRoute(r, http.MethodGet, "/api/authority/schema", s.handleAuthoritySchemaPhase7, []Format{FormatJSON})
	r.Post("/api/policy/evaluate", s.handlePolicyEvaluatePhase7)

	// Phase 8 Authority Console introspection routes
	r.Get("/api/authority/introspect", s.handleAuthorityIntrospect)
	r.Get("/api/authority/logs", s.handleAuthorityLogs)

	// Phase 9B Authority Console live routes (direct mounting instead of subrouter)
	r.Get("/api/authority/decision/{id}", authorityapi.HandleGetDecision)
	r.Get("/api/authority/decisions", authorityapi.HandleListDecisions)
	r.Post("/api/authority/decision", authorityapi.HandlePostDecision)
	r.Get("/api/authority/events", authorityapi.HandleListEventsLive)
	r.Get("/api/authority/schema/summary", authorityapi.HandleGetSchemaSummary) // WebSocket endpoint (no auth middleware due to upgrade complications)
	r.HandleFunc("/ws/authority/events", authorityapi.HandleWebSocketEvents)

	// Phase 9A mock routes (deprecated - for backward compatibility)
	r.Get("/api/authority/decision-mock/{id}", authorityapi.HandleGetDecisionMock)
	r.Get("/api/authority/events-mock", authorityapi.HandleListEvents)
	r.HandleFunc("/ws/authority/events-mock", authorityapi.HandleEventsWebSocketMock)

	// Phase 9C: Receipt Verification & Provenance Continuity
	r.Get("/api/receipts/verify/{id}", s.handleVerifyReceipt)

	// Phase 9D: Receipt Listing & Continuity Dashboard
	r.Get("/api/receipts/list", s.handleListReceipts)
	r.Get("/api/receipts/dashboard", s.handleReceiptsDashboard)
	r.Get("/dashboard/receipts", s.handleReceiptsDashboardHTML)

	// GOV-9: Authority Continuity & Receipt Lineage
	r.Get("/api/authority/lineage/{domain_id}", s.handleAuthorityLineage)
	r.Get("/api/authority/lineage/{domain_id}/verify", s.handleAuthorityLineageVerify)

	// GOV-10: Identity Provenance & Alias Receipt Integration
	r.Get("/api/identity/lineage/{actor_id}", s.handleGetIdentityLineage)
	r.Get("/api/identity/lineage/{actor_id}/verify", s.handleGetIdentityLineageVerify)

	// GOV-11: Domain-Scoped Identity Projection & Corporeal Authentication
	r.Get("/api/identity/projections/{actor_id}", s.handleGetIdentityProjections)
	r.Get("/api/domain/{domain_id}/member/{actor_id}/identity", s.handleGetDomainMemberIdentity)

	// GOV-11F: Identity Policy Viewer Integration
	r.Get("/api/policy/identity/{domainId}", s.handleGetIdentityPolicy)

	// GOV-11G: Schema-Aware Identity Policy Editing
	r.Get("/api/identity/schema/{domainId}", s.handleGetIdentitySchema)
	r.Post("/api/identity/schema/{domainId}/adopt", s.handleAdoptSchema)
	r.Post("/api/identity/policy/{domainId}", s.handleSaveIdentityPolicy)

	// GOV-11H: Domain Branching & DSCI
	r.Post("/api/domain/{id}/branch", s.handleBranchDomain)
	r.Post("/api/domain/{id}/seat/instantiate", s.handleInstantiateSeat)
	r.Get("/api/domain/{id}/branch/info", s.handleGetBranchInfo)

	// GOV-12: Alias Canon & DSCI Integration
	r.Get("/api/domain/{id}/aliases", s.HandleGetDomainAliases)
	r.Post("/api/domain/{id}/alias/relationship", s.HandleCreateRelationshipAlias)
	r.Post("/api/domain/{id}/alias/mask", s.HandleCreateMaskAlias)

	// GOV-13: Contracts Table & DSCI Contract Wiring
	r.Post("/api/domain/{id}/contracts", s.HandleCreateContract)
	r.Get("/api/domain/{id}/contracts", s.HandleGetDomainContracts)
	r.Get("/api/domain/{id}/contracts/{contractId}", s.HandleGetContract)
	r.Post("/api/domain/{id}/contracts/{contractId}/revoke", s.HandleRevokeContract)

	// CSS Inheritance: Resolved CSS with ancestor chain applied
	r.Get("/api/domain/{id}/resolved-css", s.HandleGetResolvedCSS)

	// GOV-2: Authority Console triad endpoints (read-only)
	// GOV-3: Seat mutation endpoints (write)
	if s.triadRepo != nil {
		triadHandler := authorityapi.NewTriadHandler(s.triadRepo)

		// GOV-3: Wire mutation engine if available
		if s.mutationEngine != nil {
			triadHandler.SetMutationEngine(s.mutationEngine)
		}

		// GOV-2: Read-only endpoints
		r.Get("/api/authority/triad/{identityId}", triadHandler.GetTriadByIdentity)
		r.Get("/api/authority/flow/preview", triadHandler.PreviewFlow)
		r.Post("/api/authority/flow/eval", triadHandler.EvaluateFlow)

		// GOV-3: Write endpoints
		r.Post("/api/authority/seat/transition", triadHandler.TransitionSeat)
		r.Post("/api/authority/seat/transition/batch", triadHandler.TransitionSeatBatch)

		// GOV-3: SSE event stream
		if s.eventBus != nil {
			r.Get("/api/authority/events", s.eventBus.SSEHandler)
		}
	}

	// GOV-6: Unified domain listing - Register specific paths FIRST before wildcard routes
	// Direct registration without RegisterFormatAwareRoute to avoid Chi routing conflicts
	r.Get("/api/domain/list", s.handleDomainListChi)
	r.Get("/api/domain/list/json", s.handleDomainListChi)
	r.Get("/api/domain/list/file", s.handleDomainListChi)
	s.RegisterFormatAwareRoute(r, http.MethodGet, "/api/domains", s.handleDomainListChi, []Format{FormatJSON, FormatFile})

	// Domain routes with format support (wildcard pattern must come AFTER specific paths)
	s.RegisterFormatAwareRoute(r, http.MethodGet, "/api/domain/{id}", s.handleGetDomainChi, []Format{FormatJSON, FormatFile, FormatText})

	// Domain files endpoint with format support
	s.RegisterFormatAwareRoute(r, http.MethodGet, "/api/domain/{id}/files", s.handleDomainFiles, []Format{FormatJSON, FormatText})

	// Phase 10C: Domain Schema & Policy Integration endpoints
	r.Get("/api/domain/{id}/schema", s.GetDomainSchema)
	r.Get("/api/domain/{id}/schema/json", s.GetDomainSchema)
	r.Get("/api/domain/{id}/schema/text", s.GetDomainSchema)
	r.Get("/api/domain/{id}/policy", s.GetDomainPolicy)
	r.Get("/api/domain/{id}/policy/json", s.GetDomainPolicy)
	r.Get("/api/domain/{id}/policy/text", s.GetDomainPolicy)

	// Phase 10D: Authority Console Schema Viewer Integration
	r.Get("/api/authority/schema/domain/{id}", s.GetAuthoritySchemaForDomain)
	r.Get("/api/authority/schema/domain/{id}/json", s.GetAuthoritySchemaForDomain)
	r.Get("/api/authority/schema/domain/{id}/text", s.GetAuthoritySchemaForDomain)
	r.Post("/api/authority/schema/validate", s.HandleSchemaValidation)
	r.Get("/dashboard/schema", s.HandleSchemaViewerHTML)

	// Phase 10D: Policy Continuity Integration
	r.Get("/api/policy/continuity/{domain}", s.handlePolicyContinuity)
	r.Get("/api/policy/continuity/{domain}/json", s.handlePolicyContinuity)
	r.Get("/api/policy/continuity/{domain}/text", s.handlePolicyContinuity)
	r.Get("/api/policy/continuity", s.handlePolicyContinuityGlobal)
	r.Get("/api/policy/continuity/json", s.handlePolicyContinuityGlobal)
	r.Get("/api/policy/continuity/text", s.handlePolicyContinuityGlobal)

	// Phase 10E: Policy Continuity Remediation & Visualization
	r.Post("/api/policy/continuity/remediate", s.handleRemediateOrphans)
	r.Post("/api/policy/continuity/remediate/json", s.handleRemediateOrphans)
	r.Post("/api/policy/continuity/remediate/text", s.handleRemediateOrphans)
	r.Get("/api/dashboard/continuity", s.handleContinuityDashboard)
	r.Get("/api/dashboard/continuity/json", s.handleContinuityDashboard)
	r.Get("/api/dashboard/continuity/text", s.handleContinuityDashboard)

	// Phase 10F: Continuity Lineage Proofs
	r.Post("/api/receipts/fix", s.CreateFixReceiptHandler)
	r.Post("/api/receipts/fix/json", s.CreateFixReceiptHandler)
	r.Post("/api/receipts/fix/text", s.CreateFixReceiptHandler)
	r.Get("/api/receipts/fix", s.ListFixReceiptsHandler)
	r.Get("/api/receipts/fix/json", s.ListFixReceiptsHandler)
	r.Get("/api/receipts/fix/text", s.ListFixReceiptsHandler)
	r.Get("/api/receipts/lineage/{id}", s.GetLineageProofHandler)
	r.Get("/api/receipts/lineage/{id}/json", s.GetLineageProofHandler)
	r.Get("/api/receipts/lineage/{id}/text", s.GetLineageProofHandler)

	// Phase 10K: Rego Editor Integration
	r.Get("/api/domain/{id}/policy/inherited", s.GetDomainPolicyInherited)
	r.Get("/api/domain/{id}/policy/current", s.GetDomainPolicyCurrent)
	r.Get("/api/policy/get/{id}", s.GetDomainPolicyMode) // Phase 10L: Unified hierarchical GET
	r.Post("/api/policy/validate/{id}", s.ValidateDomainPolicy)
	r.Post("/api/policy/save/{id}", s.UpdateDomainPolicy)
	r.Post("/api/policy/eval", s.HandlePolicyEval)

	// Phase S: Seats API (S0-S6)
	r.Get("/api/domain/{id}/seats", s.GetDomainSeats)
	r.Post("/api/domain/{id}/seats/appoint", s.AppointMemberSeat)
	r.Post("/api/domain/{id}/seats/{seatId}/freeze", s.FreezeSeat)
	r.Post("/api/domain/{id}/seats/{seatId}/unfreeze", s.UnfreezeSeat)
	r.Put("/api/domain/{id}/seats/{seatId}/rego", s.UpdateSeatRego)

	// GOV-7: Prime Seat Establishment for Corporeal Domains
	r.Post("/api/domain/{id}/seat/prime", s.CreatePrimeSeat)

	// MX-3.9: Domain freeze CRUD endpoints
	r.Post("/api/domain/{id}/freeze", s.handleFreezeDomain)
	r.Post("/api/domain/{id}/unfreeze", s.handleUnfreezeDomain)
	r.Post("/api/domain/{id}/freeze/override", s.handleOverrideFreezeDomain)

	// Phase 0-R.5: Atomic Corporeal + Actor-Domain Bootstrap
	r.Handle("/api/corporeal/bootstrap", handlers.CreateCorporealBootstrap(s.corporealBootstrapper))

	// Phase 10G: Cross-Domain Proof Synchronization and Verification
	r.Post("/api/receipts/proof/sync", s.handleProofSync)
	r.Post("/api/receipts/proof/sync/json", s.handleProofSync)
	r.Post("/api/receipts/proof/sync/text", s.handleProofSync)
	r.Get("/api/receipts/proof/verify/{id}", s.handleVerifyProof)
	r.Get("/api/receipts/proof/verify/{id}/json", s.handleVerifyProof)
	r.Get("/api/receipts/proof/verify/{id}/text", s.handleVerifyProof)
	r.Get("/api/federation/summary", s.handleFederationSummary)
	r.Get("/api/federation/summary/json", s.handleFederationSummary)
	r.Get("/api/federation/summary/text", s.handleFederationSummary)
	r.Post("/api/federation/trust", s.handleCreateFederationTrust)

	// Basic CRUD operations (format-aware)
	r.Post("/api/domain", s.handleCreateDomainFormatAware)
	r.Put("/api/domain/{id}", s.handleUpdateDomainFormatAware)
}
