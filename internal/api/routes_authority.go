package api

import (
	"encoding/json"
	"net/http"
)

// registerAuthorityRoutes wires all authority endpoints.
func (s *Server) registerAuthorityRoutes() {
	mux := s.router
	mux.HandleFunc("POST /api/authority/console/evaluate", s.handleAuthorityEvaluate)
	mux.HandleFunc("GET /api/authority/console/schema", s.handleAuthoritySchema)
}

// POST /api/authority/console/evaluate
func (s *Server) handleAuthorityEvaluate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON input", http.StatusBadRequest)
		return
	}

	if s.authorityConsole == nil {
		http.Error(w, "authority console not initialized", http.StatusInternalServerError)
		return
	}

	out, err := s.authorityConsole.EvaluateAuthority(ctx, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// GET /api/authority/console/schema
func (s *Server) handleAuthoritySchema(w http.ResponseWriter, r *http.Request) {
	if s.authoritySchema == nil {
		http.Error(w, "schema not loaded", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.authoritySchema)
}
