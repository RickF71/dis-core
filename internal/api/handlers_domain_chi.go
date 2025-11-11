package api

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// handleGetDomainChi handles GET /api/domain/{id} with format support
func (s *Server) handleGetDomainChi(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	domainID := chi.URLParam(r, "id")

	// Validate domain ID
	if domainID == "" || domainID == "null" || domainID == "undefined" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "missing domain id")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Get domain data using existing logic
	ctx := r.Context()
	domain, err := s.getDomainByID(ctx, domainID)
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
		ServeAsFile(w, domain)
	case FormatText:
		ServeAsText(w, domain)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleDomainListChi handles GET /api/domains with format support
func (s *Server) handleDomainListChi(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	ctx := r.Context()

	// Get domains using existing logic
	domains, err := s.getAllDomains(ctx)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusInternalServerError, "failed to fetch domains")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Format-specific response
	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, map[string]any{
			"domains": domains,
			"total":   len(domains),
		})
	case FormatFile:
		ServeAsFile(w, domains)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleDomainCSSChi handles GET /api/domain/{id}/css with format support
func (s *Server) handleDomainCSSChi(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
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

	ctx := r.Context()

	// Get CSS content using existing logic
	css, err := s.getDomainCSS(ctx, domainID)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusNotFound, "domain CSS not found")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Format-specific response
	switch format {
	case FormatFile:
		// Serve raw CSS content
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Content-Disposition", "inline; filename=\""+domainID+".css\"")
		io.WriteString(w, css)
	case FormatJSON:
		// Serve CSS metadata as JSON
		JSON(w, http.StatusOK, map[string]any{
			"domain_id":    domainID,
			"css_content":  css,
			"content_type": "text/css",
			"size":         len(css),
		})
	case FormatText:
		// Serve as plain text
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, css)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleCreateDomainFormatAware handles POST /api/domain
func (s *Server) handleCreateDomainFormatAware(w http.ResponseWriter, r *http.Request) {
	// Domain creation is JSON-only for now
	if r.Header.Get("Content-Type") != "application/json" {
		JSONError(w, http.StatusBadRequest, "content-type must be application/json")
		return
	}

	// Delegate to existing handler and convert response
	s.handleCreateDomain(w, r)
}

// handleUpdateDomainFormatAware handles PUT /api/domain/{id}
func (s *Server) handleUpdateDomainFormatAware(w http.ResponseWriter, r *http.Request) {
	// Domain updates are JSON-only for now
	if r.Header.Get("Content-Type") != "application/json" {
		JSONError(w, http.StatusBadRequest, "content-type must be application/json")
		return
	}

	// Delegate to existing handler
	s.handleUpdateDomain(w, r)
}

// handleUpdateDomainCSSFormatAware handles POST /api/domain/{id}/css
func (s *Server) handleUpdateDomainCSSFormatAware(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		JSONError(w, http.StatusBadRequest, "missing domain id")
		return
	}

	// Read CSS content
	body, err := io.ReadAll(r.Body)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	css := strings.TrimSpace(string(body))
	if css == "" {
		JSONError(w, http.StatusBadRequest, "empty CSS not allowed")
		return
	}

	ctx := r.Context()

	// Update CSS using existing logic
	err = s.updateDomainCSS(ctx, domainID, css)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to update CSS: "+err.Error())
		return
	}

	JSONOk(w, "CSS updated successfully")
}

// Helper methods that integrate with existing domain logic

// getDomainByID retrieves domain data by ID
func (s *Server) getDomainByID(ctx context.Context, domainID string) (map[string]any, error) {
	// This would integrate with existing domain retrieval logic
	// For now, return placeholder data
	return map[string]any{
		"id":     domainID,
		"name":   "Example Domain",
		"status": "active",
	}, nil
}

// getAllDomains retrieves all domains
func (s *Server) getAllDomains(ctx context.Context) ([]map[string]any, error) {
	// This would integrate with existing domain list logic
	// For now, return placeholder data
	return []map[string]any{
		{"id": "1", "name": "Domain 1"},
		{"id": "2", "name": "Domain 2"},
	}, nil
}

// getDomainCSS retrieves CSS for a domain
func (s *Server) getDomainCSS(ctx context.Context, domainID string) (string, error) {
	// This would integrate with existing CSS retrieval logic
	// For now, return placeholder CSS
	return "body { color: #fff; background: #000; }", nil
}

// updateDomainCSS updates CSS for a domain
func (s *Server) updateDomainCSS(ctx context.Context, domainID, css string) error {
	// This would integrate with existing CSS update logic
	// For now, just return success
	return nil
}

// handleDomainCSSFormatAware handles GET /api/domain/{id}/css with explicit format parameter
func (s *Server) handleDomainCSSFormatAware(w http.ResponseWriter, r *http.Request, format Format) {
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

	ctx := r.Context()

	// Get CSS content using existing logic
	css, err := s.getDomainCSS(ctx, domainID)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusNotFound, "domain CSS not found")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	switch format {
	case FormatFile:
		// Serve as CSS file download
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Content-Disposition", "attachment; filename=\"domain-"+domainID+".css\"")
		io.WriteString(w, css)
	case FormatJSON:
		// Serve CSS metadata as JSON
		JSON(w, http.StatusOK, map[string]any{
			"domain_id":    domainID,
			"css_content":  css,
			"content_type": "text/css",
			"size":         len(css),
		})
	case FormatText:
		// Serve as plain text
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, css)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}
