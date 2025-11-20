package api

import (
	"encoding/json"
	"net/http"
)

// RegisterSchemaRoutes registers all schema-related API routes.
func (s *Server) registerSchemaRoutes() {
	mux := s.router
	mux.HandleFunc("GET /api/schema/active", s.handleGetActiveSchema)
	mux.HandleFunc("GET /api/schema/list", s.handleListSchemas)
}

// GET /api/schema/active
func (s *Server) handleGetActiveSchema(w http.ResponseWriter, r *http.Request) {
	if s.authoritySchema == nil {
		http.Error(w, "schema not loaded", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(s.authoritySchema)
}

// GET /api/schema/list
func (s *Server) handleListSchemas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Ensure DB is present
	db := s.requireDB(w)
	if db == nil {
		return
	}

	rows, err := db.Query(ctx, `
		SELECT name, version, is_active
		FROM schemas
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, "failed to list schemas: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var name, version string
		var isActive bool
		if err := rows.Scan(&name, &version, &isActive); err != nil {
			http.Error(w, "scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		entry := map[string]interface{}{
			"name":      name,
			"version":   version,
			"is_active": isActive,
		}
		list = append(list, entry)
	}

	if rows.Err() != nil {
		http.Error(w, "rows error: "+rows.Err().Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		http.Error(w, "encode error: "+err.Error(), http.StatusInternalServerError)
	}
}
