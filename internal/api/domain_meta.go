package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleDomainMeta(w http.ResponseWriter, r *http.Request) {
	meta := map[string]string{
		"domain_id": "domain.user.rick",
		"display":   "Rick Personal Domain",
		"schema":    "domain.person",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meta)
}
