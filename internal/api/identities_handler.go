package api

import (
	"net/http"
)

// handleIdentities manages GET and POST requests for /identities and /api/identity/list.
// DEPRECATED: This handler uses legacy db utility functions and is not currently registered.
// Use internal/api/identities package instead.
func (s *Server) handleIdentities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	JSON(w, http.StatusNotImplemented, map[string]string{
		"error": "This handler is deprecated. Use /api/identities endpoint instead.",
	})
}
