package api

import (
	"encoding/json"
	"net/http"
	"time"
)

const (
	buildVersion = "v0.9.3"
	buildHash    = "local"
)

func (s *Server) registerVersionRoutes() {
	mux := s.mux

	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]string{
			"version":  buildVersion,
			"coreHash": buildHash,
			"time":     time.Now().UTC().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(resp)
	})
}
