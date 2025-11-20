// domain_css_bridge.go implements CSS Interchange Bridge for transparent
// conversion between CSS text, JSON, and DB objects
//
// TODO: future GUI CSS editor will consume /api/domain/{id}/css (JSON)
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"dis-core/internal/models"
	"dis-core/internal/utils"
)

// CSSValidationMiddleware validates CSS content and rejects unsafe constructs
func CSSValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only validate on PUT/POST requests
		if r.Method != http.MethodPut && r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		// Skip validation for non-CSS endpoints
		if !isCSSEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Read body for validation
		body, err := io.ReadAll(r.Body)
		if err != nil {
			JSONError(w, http.StatusBadRequest, "failed to read request body")
			return
		}

		// Create new reader for the handler
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		// Validate CSS content based on content type
		contentType := r.Header.Get("Content-Type")
		var cssContent string

		if contentType == "application/json" || strings.Contains(r.URL.Path, "/json") {
			// Parse JSON to extract CSS content
			var css models.DomainCSS
			if err := json.Unmarshal(body, &css); err != nil {
				JSONError(w, http.StatusBadRequest, "invalid JSON format")
				return
			}
			cssContent = css.CSSContent
		} else {
			// Treat as raw CSS
			cssContent = string(body)
		}

		// Perform CSS validation
		if err := utils.ValidateCSS(cssContent); err != nil {
			if cssErr, ok := err.(*models.CSSValidationError); ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(cssErr)
				return
			}
			JSONError(w, http.StatusBadRequest, fmt.Sprintf("CSS validation failed: %v", err))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isCSSEndpoint checks if the path is a CSS-related endpoint
func isCSSEndpoint(path string) bool {
	return strings.Contains(path, "/css")
}

// handleDomainCSSBridge handles all CSS operations with full interchange support
func (s *Server) handleDomainCSSBridge(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")
	if domainID == "" {
		JSONError(w, http.StatusBadRequest, "missing domain id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetDomainCSS(w, r, domainID)
	case http.MethodPut, http.MethodPost: // Phase 10J.3: Support both PUT and POST
		s.handlePutDomainCSS(w, r, domainID)
	default:
		JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleDomainCSSBridgeText handles text-specific CSS operations
func (s *Server) handleDomainCSSBridgeText(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")
	if domainID == "" {
		http.Error(w, "missing domain id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetDomainCSSText(w, r, domainID)
	case http.MethodPut, http.MethodPost: // Phase 10J.3: Support both PUT and POST
		s.handlePutDomainCSSText(w, r, domainID)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetDomainCSS handles GET /api/domain/{id}/css (JSON format by default)
// Phase 10J.4: Now queries payload->css directly
func (s *Server) handleGetDomainCSS(w http.ResponseWriter, r *http.Request, domainID string) {
	ctx := r.Context()

	db := s.requireDB(w)
	if db == nil {
		return
	}

	var cssJSON json.RawMessage
	err := db.QueryRow(ctx, `
		SELECT COALESCE(payload->'css', '{}'::jsonb)
		FROM domains
		WHERE id = $1::uuid
	`, domainID).Scan(&cssJSON)

	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "domain not found", http.StatusNotFound)
			return
		}
		JSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get CSS: %v", err))
		return
	}

	// Return JSON CSS structure
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(cssJSON)
}

// handleGetDomainCSSText handles GET /api/domain/{id}/css/text (raw CSS)
// Phase 10J.4: Now queries payload->css directly instead of domain_css table
func (s *Server) handleGetDomainCSSText(w http.ResponseWriter, r *http.Request, domainID string) {
	ctx := r.Context()

	db := s.requireDB(w)
	if db == nil {
		return
	}

	var cssContent string
	err := db.QueryRow(ctx, `
		SELECT COALESCE(payload->'css'->>'content', '')
		FROM domains
		WHERE id = $1::uuid
	`, domainID).Scan(&cssContent)

	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "domain not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to get CSS: %v", err), http.StatusInternalServerError)
		return
	}

	// Return raw CSS with proper Content-Type
	w.Header().Set("Content-Type", "text/css")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, cssContent)
}

// handlePutDomainCSS handles PUT /api/domain/{id}/css (JSON format)
func (s *Server) handlePutDomainCSS(w http.ResponseWriter, r *http.Request, domainID string) {
	ctx := r.Context()

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	// Parse JSON
	css, err := utils.CSSFromJSON(body)
	if err != nil {
		JSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}

	// Ensure domain ID matches URL parameter
	css.DomainID = domainID

	// Validate CSS content
	if err := utils.ValidateCSS(css.CSSContent); err != nil {
		// Return structured validation error
		if cssErr, ok := err.(*models.CSSValidationError); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(cssErr)
			return
		}
		JSONError(w, http.StatusBadRequest, fmt.Sprintf("CSS validation failed: %v", err))
		return
	}

	db := s.requireDB(w)
	if db == nil {
		return
	}

	// Phase 10J.4: Update payload->css directly
	_, err = db.Exec(ctx, `
		UPDATE domains
		SET payload = jsonb_set(
			payload,
			'{css}',
			jsonb_build_object(
				'content', $2::text,
				'hash', '',
				'verified', true,
				'updated_at', now()::text
			),
			true
		),
		updated_at = now()
		WHERE id = $1::uuid
	`, domainID, css.CSSContent)

	if err != nil {
		JSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to store CSS: %v", err))
		return
	}

	// Return updated CSS object
	JSON(w, http.StatusOK, css)
}

// handlePutDomainCSSText handles PUT /api/domain/{id}/css/text (raw CSS body)
// Phase 10J.4: Now updates payload->css directly instead of domain_css table
func (s *Server) handlePutDomainCSSText(w http.ResponseWriter, r *http.Request, domainID string) {
	ctx := r.Context()

	// Read raw CSS content
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	cssContent := string(body)

	// Validate CSS content
	if err := utils.ValidateCSS(cssContent); err != nil {
		// Return JSON error even for text endpoint to maintain consistency
		if cssErr, ok := err.(*models.CSSValidationError); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(cssErr)
			return
		}
		http.Error(w, fmt.Sprintf("CSS validation failed: %v", err), http.StatusBadRequest)
		return
	}

	db := s.requireDB(w)
	if db == nil {
		return
	}

	// Update payload->css directly in database
	_, err = db.Exec(ctx, `
		UPDATE domains
		SET payload = jsonb_set(
			payload,
			'{css}',
			jsonb_build_object(
				'content', $2::text,
				'hash', '',
				'verified', true,
				'updated_at', now()::text
			),
			true
		),
		updated_at = now()
		WHERE id = $1::uuid
	`, domainID, cssContent)

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to store CSS: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success with size info
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "CSS updated successfully. Size: %d bytes", len(cssContent))
}

// handleDomainCSSHistory handles GET /api/domain/{id}/css/history
func (s *Server) handleDomainCSSHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	domainID := chi.URLParam(r, "id")
	if domainID == "" {
		JSONError(w, http.StatusBadRequest, "missing domain id")
		return
	}

	// Phase 10J.4: CSS history feature removed with domain_css table decommission
	JSON(w, http.StatusOK, map[string]interface{}{
		"domain_id": domainID,
		"message":   "CSS history feature removed in Phase 10J.4 (domain_css decommission)",
		"history":   []interface{}{},
	})
}

// handleVerifyDomainCSS handles POST /api/domain/{id}/css/verify for CSS verification
func (s *Server) handleVerifyDomainCSS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	domainID := chi.URLParam(r, "id")
	if domainID == "" {
		JSONError(w, http.StatusBadRequest, "missing domain id")
		return
	}

	cssBytes, err := io.ReadAll(r.Body)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	css := strings.TrimSpace(string(cssBytes))
	if css == "" {
		JSONError(w, http.StatusBadRequest, "empty CSS body")
		return
	}

	if err := utils.ValidateCSS(css); err != nil {
		JSONError(w, http.StatusBadRequest, fmt.Sprintf("CSS validation failed: %v", err))
		return
	}

	hash := utils.CalculateCSSHash(css)
	JSON(w, http.StatusOK, map[string]interface{}{
		"verified":  true,
		"hash":      hash,
		"domain_id": domainID,
	})
}

// handleGetCSSVariables handles GET /api/domain/{id}/css/vars for CSS variable extraction
func (s *Server) handleGetCSSVariables(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	domainID := chi.URLParam(r, "id")
	if domainID == "" {
		JSONError(w, http.StatusBadRequest, "missing domain id")
		return
	}

	ctx := r.Context()

	db := s.requireDB(w)
	if db == nil {
		return
	}

	// Phase 10J.4: Get CSS from payload
	var cssContent string
	err := db.QueryRow(ctx, `
		SELECT COALESCE(payload->'css'->>'content', '')
		FROM domains
		WHERE id = $1::uuid
	`, domainID).Scan(&cssContent)

	if err != nil {
		if err == pgx.ErrNoRows {
			// Return empty variable map if domain not found
			emptyVariableMap := utils.CSSVariableMap{
				Count:     0,
				Variables: make(map[string]string),
				Hash:      utils.CalculateCSSHash(""),
				DomainID:  domainID,
			}
			JSON(w, http.StatusOK, emptyVariableMap)
			return
		}
		JSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get CSS: %v", err))
		return
	}

	// Extract CSS variables
	variableMap := utils.ExtractCSSVariableMap(cssContent, domainID)

	// Return variable map with metadata
	JSON(w, http.StatusOK, variableMap)
}

// handleUpdateDomainCSSFormatAwareBridge handles PUT with format-awareness
func (s *Server) handleUpdateDomainCSSFormatAwareBridge(w http.ResponseWriter, r *http.Request, format Format) {
	domainID := chi.URLParam(r, "id")
	if domainID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "missing domain id")
		default:
			http.Error(w, "missing domain id", http.StatusBadRequest)
		}
		return
	}

	ctx := r.Context()

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "failed to read request body")
		default:
			http.Error(w, "failed to read request body", http.StatusBadRequest)
		}
		return
	}

	var css models.DomainCSS

	// Parse based on format
	switch format {
	case FormatJSON:
		css, err = utils.CSSFromJSON(body)
		if err != nil {
			JSONError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
			return
		}
	case FormatText:
		css = utils.CSSFromText(body, domainID)
	default:
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusUnsupportedMediaType, fmt.Sprintf("unsupported format: %s", format))
		default:
			http.Error(w, fmt.Sprintf("unsupported format: %s", format), http.StatusUnsupportedMediaType)
		}
		return
	}

	// Ensure domain ID matches URL parameter
	css.DomainID = domainID

	// Validate CSS content
	if err := utils.ValidateCSS(css.CSSContent); err != nil {
		// Return structured validation error
		if cssErr, ok := err.(*models.CSSValidationError); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(cssErr)
			return
		}

		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, fmt.Sprintf("CSS validation failed: %v", err))
		default:
			http.Error(w, fmt.Sprintf("CSS validation failed: %v", err), http.StatusBadRequest)
		}
		return
	}

	// Phase 10J.4: Update payload->css directly
	db := s.requireDB(w)
	if db == nil {
		return
	}

	_, err = db.Exec(ctx, `
		UPDATE domains
		SET payload = jsonb_set(
			payload,
			'{css}',
			jsonb_build_object(
				'content', $2::text,
				'hash', '',
				'verified', true,
				'updated_at', now()::text
			),
			true
		),
		updated_at = now()
		WHERE id = $1::uuid
	`, domainID, css.CSSContent)

	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusInternalServerError, fmt.Sprintf("failed to store CSS: %v", err))
		default:
			http.Error(w, fmt.Sprintf("failed to store CSS: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Return response based on format
	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, css)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "CSS updated successfully. Size: %d bytes", css.Size)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

// registerDomainCSSBridgeRoutes registers all CSS Interchange Bridge routes
func (s *Server) registerDomainCSSBridgeRoutes() {
	// Main CSS endpoints with format support
	s.router.Route("/api/domain/{id}/css", func(r chi.Router) {
		// Default JSON format
		r.Get("/", s.handleDomainCSSBridge)
		r.Put("/", s.handleDomainCSSBridge)
		r.Post("/", s.handleDomainCSSBridge) // Phase 10J.3: Also support POST per user spec

		// Explicit text format endpoints
		r.Get("/text", s.handleDomainCSSBridgeText)
		r.Put("/text", s.handleDomainCSSBridgeText)
		r.Post("/text", s.handleDomainCSSBridgeText) // Phase 10J.3: Also support POST per user spec

		// History and verification
		r.Get("/history", s.handleDomainCSSHistory)
		r.Post("/verify", s.handleVerifyDomainCSS)

		// Phase 10I.2: CSS Variable Map Extraction
		r.Get("/vars", s.handleGetCSSVariables)
	})

	// Keep existing format-aware routes for backward compatibility
	allowedFormats := []Format{FormatFile, FormatText, FormatJSON}
	s.router.Get("/api/domain/{id}/css",
		WithFormatGuard(s.handleDomainCSSFormatAware, allowedFormats))

	// Format-specific routes
	for _, format := range allowedFormats {
		formatPattern := "/api/domain/{id}/css/" + string(format)
		s.router.Get(formatPattern,
			WithFormatGuard(s.handleDomainCSSFormatAware, allowedFormats))
	}

	// PUT routes for format-aware updates
	for _, format := range allowedFormats {
		formatPattern := "/api/domain/{id}/css/" + string(format)
		s.router.Put(formatPattern,
			WithFormatGuard(s.handleUpdateDomainCSSFormatAwareBridge, allowedFormats))
	}
}
