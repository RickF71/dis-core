package api

import (
	"net/http"

	authorityapi "dis-core/internal/api/authority"
)

// RegisterAllRoutes wires all endpoint groups into the server router with format-aware support.
// Phase 5 implementation using Chi router exclusively.
func (s *Server) RegisterAllRoutes() {
	r := s.router

	// Phase 10I: Apply CSS validation middleware to all routes
	r.Use(CSSValidationMiddleware)

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

	// Domain routes with format support
	s.RegisterFormatAwareRoute(r, http.MethodGet, "/api/domain/{id}", s.handleGetDomainChi, []Format{FormatJSON, FormatFile, FormatText})
	s.RegisterFormatAwareRoute(r, http.MethodGet, "/api/domains", s.handleDomainListChi, []Format{FormatJSON, FormatFile})

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

	// Basic CRUD operations (non-format-aware for now)
	r.Post("/api/domain", s.handleCreateDomainChi)
	r.Put("/api/domain/{id}", s.handleUpdateDomainChi)
}
