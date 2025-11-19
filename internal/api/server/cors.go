package server

import "net/http"

// WithCORS is DEPRECATED and removed
// CORS is now handled by middleware.CORSMiddleware in the middleware chain
// This function is kept for backward compatibility but does nothing
func WithCORS(next http.Handler) http.Handler {
	// CORS is handled by middleware.CORSMiddleware
	// This wrapper is no longer needed
	return next
}
