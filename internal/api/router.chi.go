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

	// Register format-specific patterns
	for _, format := range formats {
		formatPattern := pattern + "/" + string(format)
		r.Method(method, formatPattern, handler)
	}
}

// FormatAwareHandlerWrapper wraps a handler to be format-aware
func FormatAwareHandlerWrapper(handler func(w http.ResponseWriter, r *http.Request, format Format)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := DetectFormat(r)

		// Strip format suffix from path for processing
		originalPath := r.URL.Path
		r.URL.Path = StripFormatSuffix(originalPath)
		defer func() { r.URL.Path = originalPath }()

		handler(w, r, format)
	}
}
