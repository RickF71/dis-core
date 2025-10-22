package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) registerNetworkRoutes() {
	mux := s.mux

	mux.HandleFunc("/api/net/peers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			peers := []string{"localhost"} // placeholder list
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"count": len(peers),
				"peers": peers,
			})
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

}
