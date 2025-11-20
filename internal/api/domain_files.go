package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"dis-core/internal/db"

	"github.com/go-chi/chi/v5"
)

// handleDomainFiles handles GET /api/domain/{id}/files with format support
func (s *Server) handleDomainFiles(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "Domain ID required")
		case FormatText:
			http.Error(w, "Domain ID required", http.StatusBadRequest)
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	ctx := r.Context()

	// Require DB
	dbPool := s.requireDB(w)
	if dbPool == nil {
		return
	}

	// Get files using the DB helper
	fileNames, err := db.ListFilesForDomain(ctx, dbPool, domainID)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusNotFound, "Domain not found")
		case FormatText:
			http.Error(w, "Domain not found", http.StatusNotFound)
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Format-specific response
	switch format {
	case FormatJSON:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fileNames)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain")
		if len(fileNames) > 0 {
			response := strings.Join(fileNames, "\n")
			w.Write([]byte(response))
		} else {
			w.Write([]byte(""))
		}
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleDomainFilesFormatAware handles GET /api/domain/{id}/files with explicit format parameter
func (s *Server) handleDomainFilesFormatAware(w http.ResponseWriter, r *http.Request, format Format) {
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "Domain ID required")
		case FormatText:
			http.Error(w, "Domain ID required", http.StatusBadRequest)
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	ctx := r.Context()

	dbPool := s.requireDB(w)
	if dbPool == nil {
		return
	}

	// Get files using the DB helper
	fileNames, err := db.ListFilesForDomain(ctx, dbPool, domainID)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusNotFound, "Domain not found")
		case FormatText:
			http.Error(w, "Domain not found", http.StatusNotFound)
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Format-specific response
	switch format {
	case FormatJSON:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fileNames)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain")
		if len(fileNames) > 0 {
			response := strings.Join(fileNames, "\n")
			w.Write([]byte(response))
		} else {
			w.Write([]byte(""))
		}
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// Legacy handler for backward compatibility
func (s *Server) ListDomainFiles(w http.ResponseWriter, r *http.Request) {
	s.handleDomainFiles(w, r)
}
