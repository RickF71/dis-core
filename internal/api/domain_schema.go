package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
)

// GetDomainSchema handles GET /api/domain/{id}/schema with format support
func (s *Server) GetDomainSchema(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "Domain ID required")
		case FormatText:
			http.Error(w, "Domain ID required", http.StatusBadRequest)
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	ctx := r.Context()

	// Try to get schema from database first (only if database is available).
	// Use s.DB() (nil-safe) rather than requireDB which writes a 503 response:
	// we want to fall back to filesystem when DB isn't configured.
	var schema map[string]interface{}
	var err error

	db := s.DB()
	if db != nil {
		schema, err = s.getDomainSchemaFromDB(ctx, db, domainID)
	}

	if db == nil || err != nil || schema == nil {
		// Fallback: try filesystem
		schema = s.getDomainSchemaFromFS(domainID)
	} // Always return 200 OK even if no schema found (return empty object)
	if schema == nil {
		schema = map[string]interface{}{}
	}

	// Format-specific response
	switch format {
	case FormatJSON:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(schema)
	case FormatText:
		w.Header().Set("Content-Type", "text/plain")
		yamlContent, err := yaml.Marshal(schema)
		if err != nil {
			// Fallback to JSON-like text representation
			jsonBytes, _ := json.MarshalIndent(schema, "", "  ")
			w.Write(jsonBytes)
		} else {
			w.Write(yamlContent)
		}
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// getDomainSchemaFromDB queries the domain_schemas table for the schema
// getDomainSchemaFromDB queries the domain_schemas table for the schema
func (s *Server) getDomainSchemaFromDB(ctx context.Context, db *pgxpool.Pool, domainID string) (map[string]interface{}, error) {
	var schemaContent []byte
	query := `SELECT schema_content FROM domain_schemas WHERE domain_id = $1`

	if db == nil {
		return nil, fmt.Errorf("database not available")
	}

	err := db.QueryRow(ctx, query, domainID).Scan(&schemaContent)
	if err != nil {
		// Schema not found in database
		return nil, err
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(schemaContent, &schema); err != nil {
		return nil, fmt.Errorf("invalid schema JSON: %w", err)
	}

	return schema, nil
}

// getDomainSchemaFromFS tries to read schema from filesystem
func (s *Server) getDomainSchemaFromFS(domainID string) map[string]interface{} {
	// Try multiple possible paths
	paths := []string{
		fmt.Sprintf("data/domains/%s/schema.json", domainID),
		fmt.Sprintf("internal/domains/%s/schema.json", domainID),
		fmt.Sprintf("domains/%s/schema.json", domainID),
	}

	for _, path := range paths {
		if content, err := os.ReadFile(path); err == nil {
			var schema map[string]interface{}
			if json.Unmarshal(content, &schema) == nil {
				return schema
			}
		}
	}

	// Return nil if no schema found
	return nil
}
