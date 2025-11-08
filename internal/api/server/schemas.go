package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleSchemaList(w http.ResponseWriter, r *http.Request) {
	if s.Schemas == nil {
		http.Error(w, "schema registry unavailable", http.StatusServiceUnavailable)
		return
	}

	type schemaInfo struct {
		ID      string `json:"id"`
		Version string `json:"version"`
	}
	var out []schemaInfo
	for _, e := range s.Schemas.ByKey() {
		out = append(out, schemaInfo{ID: e.ID, Version: e.Version})
	}
	json.NewEncoder(w).Encode(out)
}
