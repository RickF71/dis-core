package api

import (
	"encoding/json"
	"net/http"
)

// registerAuthorityRoutes wires all authority endpoints.
func (s *Server) registerAuthorityRoutes() {
	r := s.router
	// Console evaluate/schema endpoints (legacy console)
	r.Post("/api/authority/console/evaluate", s.handleAuthorityEvaluate)
	r.Get("/api/authority/console/schema", s.handleAuthoritySchema)

	// New core authority delegators (MX-3.5)
	if s.Engine != nil {
		r.Post("/api/authority/freeze", NewFreezeHandler(s.Engine))
		r.Post("/api/authority/unfreeze", NewUnfreezeHandler(s.Engine))
		r.Post("/api/authority/override", NewOverrideHandler(s.Engine))
		r.Get("/api/authority/status", HandleAuthorityStatusHandler(s.Engine))
		r.Get("/api/authority/introspect", HandleAuthorityIntrospectHandler(s.Engine))
		r.Get("/api/authority/lineage", HandleAuthorityLineageHandler(s.Engine))

		// Seat lifecycle endpoints (MX-3.6)
		r.Post("/api/authority/seat/instantiate", NewSeatInstantiateHandler(s.Engine))
		r.Get("/api/authority/seat/{id}/status", NewSeatStatusHandler(s.Engine))
		r.Get("/api/authority/seat/{id}/lineage", NewSeatLineageHandler(s.Engine))
		r.Get("/api/authority/domain/{id}/seats", NewSeatListHandler(s.Engine))
	} else {
		// If engine not wired, register placeholders that return 503
		r.Post("/api/authority/freeze", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
		})
		r.Post("/api/authority/unfreeze", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
		})
		r.Post("/api/authority/override", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
		})
		r.Get("/api/authority/status", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
		})
		r.Get("/api/authority/introspect", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
		})
		r.Get("/api/authority/lineage", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
		})
	}
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
