package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewChiRouter creates the main chi router with all format-aware routes
func (s *Server) NewChiRouter() *chi.Mux {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Register all format-aware routes
	s.registerCoreRoutes(r)
	s.registerDomainRoutes(r)
	s.registerPolicyRoutes(r)
	s.registerStatusRoutes(r)
	s.registerFileRoutesForChi(r)
	s.registerReceiptRoutesForChi(r)

	return r
}

// registerCoreRoutes registers basic API endpoints
func (s *Server) registerCoreRoutes(r *chi.Mux) {
	// Health and connectivity endpoints with JSON/text support
	s.RegisterFormatAwareRoute(r, "GET", "/api/ping", s.handlePingFormatAware,
		[]Format{FormatJSON, FormatText})
	s.RegisterFormatAwareRoute(r, "GET", "/api/health", s.handleHealthFormatAware,
		[]Format{FormatJSON, FormatText})
}

// registerDomainRoutes registers domain-related endpoints
func (s *Server) registerDomainRoutes(r *chi.Mux) {
	// Domain CRUD with JSON/file support
	s.RegisterFormatAwareRoute(r, "GET", "/api/domain/{id}", s.handleGetDomainChiFormatAware,
		[]Format{FormatJSON, FormatFile, FormatText})
	s.RegisterFormatAwareRoute(r, "GET", "/api/domains", s.handleDomainListChiFormatAware,
		[]Format{FormatJSON, FormatFile})

	// Domain creation (JSON only)
	r.Post("/api/domain", s.handleCreateDomainChiFormatAware)
	r.Put("/api/domain/{id}", s.handleUpdateDomainChiFormatAware)

	// Domain CSS with file/JSON support
	s.RegisterFormatAwareRoute(r, "GET", "/api/domain/{id}/css", s.handleDomainCSSChiFormatAware,
		[]Format{FormatFile, FormatJSON, FormatText})
	r.Post("/api/domain/{id}/css", s.handleUpdateDomainCSSChiFormatAware)

	// Domain files with format awareness
	s.RegisterFormatAwareRoute(r, "GET", "/api/domain/{id}/files", s.handleDomainFilesListChiFormatAware,
		[]Format{FormatJSON, FormatFile})
	s.RegisterFormatAwareRoute(r, "GET", "/api/domain/{id}/file/{filename}", s.handleDomainFileGetChiFormatAware,
		[]Format{FormatFile, FormatJSON})

	// Domain policy with file/JSON support
	s.RegisterFormatAwareRoute(r, "GET", "/api/domain/{id}/policy", s.handleGetDomainPolicyChiFormatAware,
		[]Format{FormatFile, FormatJSON, FormatText})
	r.Post("/api/domain/{id}/policy", s.handleSetDomainPolicyChiFormatAware)

	// Other domain endpoints (file operations)
	r.Put("/api/domain/{id}/file/{filename}", s.handleDomainFilePutChiFormatAware)
	r.Delete("/api/domain/{id}/file/{filename}", s.handleDomainFileDeleteChiFormatAware)
	r.Post("/api/domain/{id}/file/{filename}", s.handleDomainFileCreateChiFormatAware)
	r.Post("/api/domain/{id}/file/rename", s.handleDomainFileRenameChiFormatAware)

	// Domain announcements with text/JSON support
	s.RegisterFormatAwareRoute(r, "GET", "/api/domain/{id}/announce", s.handleDomainAnnounceChiFormatAware,
		[]Format{FormatJSON, FormatText})
} // registerPolicyRoutes registers policy management endpoints
func (s *Server) registerPolicyRoutes(r *chi.Mux) {
	// Policy list with JSON/file support
	s.RegisterFormatAwareRoute(r, "GET", "/api/policy/list", s.handleListPoliciesFormatAware,
		[]Format{FormatJSON, FormatFile})

	// Individual policy with file/JSON support
	s.RegisterFormatAwareRoute(r, "GET", "/api/policy/{name}", s.handleGetPolicyFormatAware,
		[]Format{FormatFile, FormatJSON, FormatText})

	// Policy CRUD operations
	r.Put("/api/policy/{name}", s.handlePutPolicyFormatAware)
	r.Delete("/api/policy/{name}", s.handleDeletePolicyFormatAware)
	r.Post("/api/policy/reload", s.handlePolicyReloadFormatAware)
}

// registerStatusRoutes registers status and monitoring endpoints
func (s *Server) registerStatusRoutes(r *chi.Mux) {
	// Status endpoint with JSON/text support
	s.RegisterFormatAwareRoute(r, "GET", "/api/status", s.handleStatusFormatAware,
		[]Format{FormatJSON, FormatText})

	// Authority status with JSON/text support
	s.RegisterFormatAwareRoute(r, "GET", "/api/authority/status", s.handleAuthorityStatusFormatAware,
		[]Format{FormatJSON, FormatText})
}

// registerFileRoutesForChi registers file management endpoints for chi router
func (s *Server) registerFileRoutesForChi(r *chi.Mux) {
	// File search and export with JSON/file support
	s.RegisterFormatAwareRoute(r, "GET", "/api/files/search", s.handleFileSearchFormatAware,
		[]Format{FormatJSON, FormatFile})
	s.RegisterFormatAwareRoute(r, "GET", "/api/files/export", s.handleFileExportFormatAware,
		[]Format{FormatJSON, FormatFile})
	s.RegisterFormatAwareRoute(r, "GET", "/api/files", s.handleListFilesFormatAware,
		[]Format{FormatJSON, FormatFile})
}

// registerReceiptRoutesForChi registers receipt management endpoints for chi router
func (s *Server) registerReceiptRoutesForChi(r *chi.Mux) {
	// Receipt listing with JSON/file/text support
	s.RegisterFormatAwareRoute(r, "GET", "/api/receipts", s.handleListReceiptsFormatAware,
		[]Format{FormatJSON, FormatFile, FormatText})

	// Receipt search with JSON/file/text support
	s.RegisterFormatAwareRoute(r, "GET", "/api/receipts/search", s.handleReceiptSearchFormatAware,
		[]Format{FormatJSON, FormatFile, FormatText})

	// Individual receipt with file/JSON/text support
	s.RegisterFormatAwareRoute(r, "GET", "/api/receipt/{id}", s.handleGetReceiptFormatAware,
		[]Format{FormatFile, FormatJSON, FormatText})

	// Receipt verification with JSON/text support
	s.RegisterFormatAwareRoute(r, "GET", "/api/receipt/{id}/verify", s.handleVerifyReceiptFormatAware,
		[]Format{FormatJSON, FormatText})

	// Receipt CRUD operations
	r.Post("/api/receipt", s.handleCreateReceiptFormatAware)
	r.Delete("/api/receipt/{id}", s.handleDeleteReceiptFormatAware)

	// Version management with JSON/file/text support
	s.RegisterFormatAwareRoute(r, "GET", "/api/versions", s.handleListVersionsFormatAware,
		[]Format{FormatJSON, FormatFile, FormatText})
	s.RegisterFormatAwareRoute(r, "GET", "/api/version/{version}/export", s.handleVersionExportFormatAware,
		[]Format{FormatJSON, FormatFile})
}
