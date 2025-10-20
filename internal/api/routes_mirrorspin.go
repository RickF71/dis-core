package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/mirrorspin"
)

// registerMirrorSpinRoutes exposes diagnostic endpoints for the MirrorSpin engine.
func (s *Server) registerMirrorSpinRoutes() {
	mux := s.mux

	// GET /api/mirrorspin/status
	mux.HandleFunc("/api/mirrorspin/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := mirrorspin.GetStatus()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})
}
