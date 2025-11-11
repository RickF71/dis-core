package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewFormatAwareRouter creates a chi router with format suffix support
func NewFormatAwareRouter() *chi.Mux {
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	return r
}

// RegisterFormatAwareRoute registers a handler for multiple format suffixes
func (s *Server) RegisterFormatAwareRoute(r *chi.Mux, method, pattern string, handler http.HandlerFunc, formats []Format) {
	// Register base pattern (defaults to JSON)
	r.Method(method, pattern, handler)

	// Register format-specific patterns for all supported formats (including unimplemented ones)
	allFormats := []Format{FormatJSON, FormatText, FormatFile, FormatCBOR, FormatEncrypted}
	for _, format := range allFormats {
		formatPattern := pattern + "/" + string(format)
		r.Method(method, formatPattern, handler)
	}
}

func FormatAwareHandlerWrapper(handler func(w http.ResponseWriter, r *http.Request, format Format)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := DetectFormat(r)

		// Strip only the known suffix, not query or parameters
		originalPath := r.URL.Path
		r.URL.Path = StripFormatSuffix(originalPath)
		defer func() { r.URL.Path = originalPath }()

		handler(w, r, format)
	}
}

// WithFormatGuard creates a format-guarded handler that validates allowed formats
func WithFormatGuard(handler func(w http.ResponseWriter, r *http.Request, format Format), allowed []Format) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := DetectFormat(r)

		// Check if the detected format is allowed
		if !isAllowedFormat(format, allowed) {
			JSONError(w, http.StatusBadRequest, "format '"+string(format)+"' not supported for this endpoint")
			return
		}

		// Strip format suffix and restore after execution
		originalPath := r.URL.Path
		r.URL.Path = StripFormatSuffix(originalPath)
		defer func() { r.URL.Path = originalPath }()

		handler(w, r, format)
	}
}

// isAllowedFormat checks if a format is in the allowed list
func isAllowedFormat(f Format, allowed []Format) bool {
	for _, allowedFormat := range allowed {
		if f == allowedFormat {
			return true
		}
	}
	return false
}
