package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"dis-core/internal/api/middleware"
)

// GET /api/authority/introspect
func (s *Server) handleAuthorityIntrospect(w http.ResponseWriter, r *http.Request) {
	// Get policy engine from middleware
	policyEngine := middleware.FromPolicyEngine(r.Context())

	result := map[string]any{
		"policies": getPolicyModules(policyEngine),
		"schema":   getActiveSchemaInfo(s),
		"status":   "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GET /api/authority/logs
func (s *Server) handleAuthorityLogs(w http.ResponseWriter, r *http.Request) {
	logsDir := "internal/logs/phases"
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logs := []map[string]string{}
	for _, e := range entries {
		if !e.IsDir() {
			data, err := os.ReadFile(filepath.Join(logsDir, e.Name()))
			if err != nil {
				continue // Skip files that can't be read
			}
			logs = append(logs, map[string]string{
				"file":    e.Name(),
				"content": string(data),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// GET /api/authority/events
func (s *Server) handleAuthorityEvents(w http.ResponseWriter, r *http.Request) {
	db := middleware.FromDB(r.Context())
	if db == nil {
		http.Error(w, "Database unavailable", http.StatusInternalServerError)
		return
	}

	// Try receipts table, fall back to canon table if receipts doesn't exist
	rows, err := db.Query(context.Background(),
		"SELECT id, type, domain, actor, created_at FROM receipts ORDER BY created_at DESC LIMIT 50")

	if err != nil {
		// Fallback: try to get events from canon table
		rows, err = db.Query(context.Background(),
			"SELECT type, content, '00000000-0000-0000-0000-000000000000' as domain, 'system' as actor, NOW() as created_at FROM canon ORDER BY type DESC LIMIT 50")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	defer rows.Close()

	events := []map[string]any{}
	for rows.Next() {
		var id, typ, domain, actor string
		var created time.Time

		err := rows.Scan(&id, &typ, &domain, &actor, &created)
		if err != nil {
			continue // Skip rows that can't be scanned
		}

		events = append(events, map[string]any{
			"id":         id,
			"type":       typ,
			"domain":     domain,
			"actor":      actor,
			"created_at": created,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// Helper function to get policy modules info
func getPolicyModules(policyEngine interface{}) []string {
	if policyEngine == nil {
		return []string{}
	}

	// Return basic policy module list
	return []string{"gates.rego", "risk.rego", "freeze.rego"}
}

// Helper function to get active schema info
func getActiveSchemaInfo(s *Server) map[string]any {
	return map[string]any{
		"loaded_schemas": 3,
		"schema_types":   []string{"authority.layer", "schema.json", "seat.unified.v1"},
		"version":        "v1.0-dev",
	}
}
