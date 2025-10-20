package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) registerNetworkRoutes() {
	mux := s.mux

	mux.HandleFunc("/api/net/peers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			peers := s.NetManager.ListPeers()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"count": len(peers),
				"peers": peers,
			})
			return
		}

		if r.Method == http.MethodPost {
			var payload struct{ Address string }
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if payload.Address == "" {
				http.Error(w, "missing address", http.StatusBadRequest)
				return
			}

			// Add to memory
			s.NetManager.AddPeer(payload.Address)

			// Persist to DB
			if err := s.NetManager.SavePeerToDB(s.store, payload.Address); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
		}

	})
}
