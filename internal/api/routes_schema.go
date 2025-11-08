package api

import (
	"context"
	"encoding/json"
	"net/http"

	"dis-core/internal/schema"
)

// RegisterSchemaRoutes registers all schema-related API routes.
func (s *Server) registerSchemaRoutes() {
	mux := s.mux
	mux.HandleFunc("GET /api/schema/active", s.handleGetActiveSchema)
	mux.HandleFunc("GET /api/schema/list", s.handleListSchemas)
}

// GET /api/schema/active
func (s *Server) handleGetActiveSchema(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	active, err := schema.GetActiveSchema(ctx, s.db)
	if err != nil {
		http.Error(w, "failed to get active schema: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"name":      active.Name,
		"version":   active.Version,
		"is_active": active.IsActive,
		"hash":      active.Hash,
		"schema":    json.RawMessage(active.JSON),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET /api/schema/list
func (s *Server) handleListSchemas(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	rows, err := s.db.QueryContext(ctx, `
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}
