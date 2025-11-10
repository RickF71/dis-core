package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterFormatAwareRoutes registers routes that support format suffixes
func (s *Server) RegisterFormatAwareRoutes(r *chi.Mux) {
	// Basic API endpoints with JSON and text support
	s.RegisterFormatAwareRoute(r, "GET", "/api/ping", s.handlePing,
		[]Format{FormatJSON, FormatText})
	s.RegisterFormatAwareRoute(r, "GET", "/api/health", s.handleHealth,
		[]Format{FormatJSON, FormatText})

	// Status endpoint with JSON and text support
	s.RegisterFormatAwareRoute(r, "GET", "/api/status", s.HandleStatus,
		[]Format{FormatJSON, FormatText})

	// Domain endpoints with JSON and file export support
	s.RegisterFormatAwareRoute(r, "GET", "/api/domain/{id}", s.handleGetDomainFormatAware,
		[]Format{FormatJSON, FormatFile})
	s.RegisterFormatAwareRoute(r, "GET", "/api/domains", s.handleDomainListFormatAware,
		[]Format{FormatJSON, FormatFile})

	// CSS endpoint with file (raw CSS) and JSON (metadata) support
	s.RegisterFormatAwareRoute(r, "GET", "/api/domain/{id}/css", s.handleDomainCSSFormatAware,
		[]Format{FormatFile, FormatJSON})

	// Policy endpoints with file (raw policy) and JSON (metadata) support
	s.RegisterFormatAwareRoute(r, "GET", "/api/policy/{name}", s.handleGetPolicyFormatAware,
		[]Format{FormatFile, FormatJSON})
}

// Example format-aware domain handler
func (s *Server) handleGetDomainFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)

	// Get domain ID from URL (after stripping format suffix)
	domainID := chi.URLParam(r, "id")
	if domainID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "missing domain id")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Get domain data (placeholder)
	ctx := r.Context()
	domain, err := s.getDomainData(ctx, domainID)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusNotFound, "domain not found")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Format-specific response
	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, domain)
	case FormatFile:
		ServeAsFile(w, domain) // Download domain data as JSON file
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// Placeholder helper method (would integrate with existing domain logic)
func (s *Server) getDomainData(ctx any, domainID string) (map[string]any, error) {
	// This would integrate with existing domain retrieval logic
	return map[string]any{
		"domain_id": domainID,
		"name":      "example",
		"status":    "active",
	}, nil
}

// Format-aware domain list handler
func (s *Server) handleDomainListFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)

	// Get domains (placeholder)
	domains := []map[string]any{
		{"id": "1", "name": "example1"},
		{"id": "2", "name": "example2"},
	}

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, map[string]any{
			"domains": domains,
			"total":   len(domains),
		})
	case FormatFile:
		ServeAsFile(w, domains) // Export as CSV or JSON file
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// Format-aware CSS handler
func (s *Server) handleDomainCSSFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	domainID := chi.URLParam(r, "id")

	// Get CSS data (placeholder)
	cssContent := "body { color: #fff; }"

	switch format {
	case FormatFile:
		// Serve raw CSS content
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Write([]byte(cssContent))
	case FormatJSON:
		// Serve CSS metadata as JSON
		JSON(w, http.StatusOK, map[string]any{
			"domain_id":    domainID,
			"css_content":  cssContent,
			"content_type": "text/css",
			"size":         len(cssContent),
		})
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// Format-aware policy handler
func (s *Server) handleGetPolicyFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	policyName := chi.URLParam(r, "name")

	// Get policy content (placeholder)
	policyContent := "package example\n\ndefault allow = false"

	switch format {
	case FormatFile:
		// Serve raw policy content
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(policyContent))
	case FormatJSON:
		// Serve policy metadata as JSON
		JSON(w, http.StatusOK, map[string]any{
			"policy_name": policyName,
			"content":     policyContent,
			"type":        "rego",
			"size":        len(policyContent),
		})
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}
