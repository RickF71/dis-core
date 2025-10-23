package api

import (
	"encoding/json"
	"net/http"
)

// registerDBRoutes registers routes related to database health and introspection.
func (s *Server) registerDBRoutes() {
	mux := s.mux

	mux.HandleFunc("/api/db/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status, err := s.Ledger.GetDBStatus()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})
}
