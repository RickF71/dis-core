package api

import (
	"net/http"
)

// RegisterAllRoutes wires all endpoint groups into the server router with format-aware support.
// Phase 5 implementation using Chi router exclusively.
func (s *Server) RegisterAllRoutes() {
	r := s.router

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

	// Domain routes with format support
	s.RegisterFormatAwareRoute(r, http.MethodGet, "/api/domain/{id}", s.handleGetDomainChi, []Format{FormatJSON, FormatFile, FormatText})
	s.RegisterFormatAwareRoute(r, http.MethodGet, "/api/domains", s.handleDomainListChi, []Format{FormatJSON, FormatFile})

	// Basic CRUD operations (non-format-aware for now)
	r.Post("/api/domain", s.handleCreateDomainChi)
	r.Put("/api/domain/{id}", s.handleUpdateDomainChi)
}
