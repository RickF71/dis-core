package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) registerNetworkRoutes() {
	mux := s.mux

	mux.HandleFunc("/api/net/peers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			peers := s.NetManager.ListPeers()
			writeJSON(w, http.StatusOK, map[string]any{
				"count": len(peers),
				"peers": peers,
			})
			return
		case http.MethodPost:
			var payload struct {
				Address string `json:"address"`
			}
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

			// Persist to DB if store is available
			if s.store != nil {
				if err := s.NetManager.SavePeerToDB(s.store, payload.Address); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			w.WriteHeader(http.StatusCreated)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})
}
