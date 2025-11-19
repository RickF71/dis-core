package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Schema represents a registered identity policy schema
type Schema struct {
	ID        string                 `json:"id"`
	Version   string                 `json:"version"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
}

// DomainSchema represents a schema adopted by a domain
type DomainSchema struct {
	DomainID       string                 `json:"domain_id"`
	SchemaID       string                 `json:"schema_id"`
	SchemaVersion  string                 `json:"schema_version"`
	Mode           string                 `json:"mode"` // "adopted" or "inherited"
	SourceDomainID string                 `json:"source_domain_id"`
	Depth          int                    `json:"inheritance_depth"`
	AdoptedAt      time.Time              `json:"adopted_at"`
	AdoptedBy      *string                `json:"adopted_by,omitempty"`
	Payload        map[string]interface{} `json:"schema_payload"`
}

// SchemaSet represents the effective set of schemas for a domain
type SchemaSet struct {
	DomainID  string                 `json:"domain_id"`
	Schemas   []DomainSchema         `json:"schemas"`
	Effective map[string]interface{} `json:"effective_schema"` // Merged field definitions
}

// SchemaResolver handles schema loading and resolution
type SchemaResolver struct {
	db *pgxpool.Pool
}

// NewSchemaResolver creates a new schema resolver
func NewSchemaResolver(db *pgxpool.Pool) *SchemaResolver {
	return &SchemaResolver{db: db}
}

// LoadSchema loads a specific schema by ID and version
func (sr *SchemaResolver) LoadSchema(ctx context.Context, id, version string) (*Schema, error) {
	var schema Schema
	var payloadJSON []byte

	err := sr.db.QueryRow(ctx, `
		SELECT id, version, payload, created_at
		FROM schemas
		WHERE id = $1 AND version = $2
	`, id, version).Scan(&schema.ID, &schema.Version, &payloadJSON, &schema.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("schema not found: %w", err)
	}

	if err := json.Unmarshal(payloadJSON, &schema.Payload); err != nil {
		return nil, fmt.Errorf("invalid schema payload: %w", err)
	}

	return &schema, nil
}

// LoadDomainSchemas loads schemas directly adopted by a domain
func (sr *SchemaResolver) LoadDomainSchemas(ctx context.Context, domainID string) ([]DomainSchema, error) {
	rows, err := sr.db.Query(ctx, `
		SELECT
			ds.domain_id,
			ds.schema_id,
			ds.schema_version,
			'adopted' AS mode,
			ds.domain_id AS source_domain_id,
			0 AS depth,
			ds.adopted_at,
			ds.adopted_by,
			s.payload
		FROM domain_schemas ds
		JOIN schemas s ON s.id = ds.schema_id AND s.version = ds.schema_version
		WHERE ds.domain_id = $1
		ORDER BY ds.adopted_at
	`, domainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return sr.scanDomainSchemas(rows)
}

// LoadInheritedSchemas loads schemas inherited from parent chain
func (sr *SchemaResolver) LoadInheritedSchemas(ctx context.Context, domainID string) ([]DomainSchema, error) {
	rows, err := sr.db.Query(ctx, `
		WITH RECURSIVE parent_chain AS (
			-- Start with domain's parent
			SELECT
				$1::uuid AS domain_id,
				d.parent_id AS current_domain_id,
				d.parent_id,
				1 AS depth
			FROM domains d
			WHERE d.id = $1 AND d.parent_id IS NOT NULL

			UNION ALL

			-- Walk up parent chain
			SELECT
				pc.domain_id,
				d.id AS current_domain_id,
				d.parent_id,
				pc.depth + 1 AS depth
			FROM parent_chain pc
			JOIN domains d ON d.id = pc.parent_id
			WHERE pc.parent_id IS NOT NULL AND pc.depth < 10
		)
		SELECT DISTINCT ON (ds.schema_id, ds.schema_version)
			pc.domain_id,
			ds.schema_id,
			ds.schema_version,
			'inherited' AS mode,
			ds.domain_id AS source_domain_id,
			pc.depth,
			ds.adopted_at,
			ds.adopted_by,
			s.payload
		FROM parent_chain pc
		JOIN domain_schemas ds ON ds.domain_id = pc.current_domain_id
		JOIN schemas s ON s.id = ds.schema_id AND s.version = ds.schema_version
		ORDER BY ds.schema_id, ds.schema_version, pc.depth
	`, domainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return sr.scanDomainSchemas(rows)
}

// ResolveSchemaSet resolves the complete schema set for a domain (adopted + inherited)
func (sr *SchemaResolver) ResolveSchemaSet(ctx context.Context, domainID string) (*SchemaSet, error) {
	// Load adopted schemas
	adopted, err := sr.LoadDomainSchemas(ctx, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to load adopted schemas: %w", err)
	}

	// Load inherited schemas
	inherited, err := sr.LoadInheritedSchemas(ctx, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to load inherited schemas: %w", err)
	}

	// Combine (adopted first, then inherited)
	allSchemas := append(adopted, inherited...)

	// Compute effective schema
	effectiveSchema := sr.ComputeEffectiveSchema(allSchemas)

	return &SchemaSet{
		DomainID:  domainID,
		Schemas:   allSchemas,
		Effective: effectiveSchema,
	}, nil
}

// ComputeEffectiveSchema merges all schema fields into a single effective schema
// Fields from adopted schemas take precedence over inherited ones
// Within each mode, earlier schemas take precedence
func (sr *SchemaResolver) ComputeEffectiveSchema(schemas []DomainSchema) map[string]interface{} {
	effective := make(map[string]interface{})

	// Track which fields have been defined to handle precedence
	definedFields := make(map[string]bool)

	// Process schemas in order (adopted first, then inherited by depth)
	for _, schema := range schemas {
		properties, ok := schema.Payload["properties"].(map[string]interface{})
		if !ok {
			continue
		}

		for fieldName, fieldDef := range properties {
			// Only add field if not already defined by higher-precedence schema
			if !definedFields[fieldName] {
				effective[fieldName] = fieldDef
				definedFields[fieldName] = true
			}
		}
	}

	// Wrap in JSON Schema structure
	result := map[string]interface{}{
		"type":       "object",
		"properties": effective,
	}

	// Collect required fields from all schemas
	requiredFields := make(map[string]bool)
	for _, schema := range schemas {
		if required, ok := schema.Payload["required"].([]interface{}); ok {
			for _, field := range required {
				if fieldName, ok := field.(string); ok {
					requiredFields[fieldName] = true
				}
			}
		}
	}

	if len(requiredFields) > 0 {
		required := make([]string, 0, len(requiredFields))
		for field := range requiredFields {
			required = append(required, field)
		}
		result["required"] = required
	}

	return result
}

// scanDomainSchemas is a helper to scan rows into DomainSchema slice
func (sr *SchemaResolver) scanDomainSchemas(rows interface {
	Next() bool
	Scan(...interface{}) error
}) ([]DomainSchema, error) {
	var schemas []DomainSchema

	for rows.Next() {
		var ds DomainSchema
		var payloadJSON []byte
		var adoptedBy *string

		err := rows.Scan(
			&ds.DomainID,
			&ds.SchemaID,
			&ds.SchemaVersion,
			&ds.Mode,
			&ds.SourceDomainID,
			&ds.Depth,
			&ds.AdoptedAt,
			&adoptedBy,
			&payloadJSON,
		)
		if err != nil {
			return nil, err
		}

		ds.AdoptedBy = adoptedBy

		if err := json.Unmarshal(payloadJSON, &ds.Payload); err != nil {
			return nil, fmt.Errorf("invalid schema payload: %w", err)
		}

		schemas = append(schemas, ds)
	}

	return schemas, nil
}

// CheckSchemaCompatibility checks if a schema can be adopted without conflicts
func (sr *SchemaResolver) CheckSchemaCompatibility(ctx context.Context, domainID, schemaID, schemaVersion string) error {
	// Load current schema set
	currentSet, err := sr.ResolveSchemaSet(ctx, domainID)
	if err != nil {
		return fmt.Errorf("failed to resolve current schema set: %w", err)
	}

	// Load proposed schema
	proposedSchema, err := sr.LoadSchema(ctx, schemaID, schemaVersion)
	if err != nil {
		return fmt.Errorf("proposed schema not found: %w", err)
	}

	// Check if already adopted
	for _, schema := range currentSet.Schemas {
		if schema.SchemaID == schemaID && schema.SchemaVersion == schemaVersion && schema.Mode == "adopted" {
			return fmt.Errorf("schema %s@%s already adopted", schemaID, schemaVersion)
		}
	}

	// Check for field conflicts (optional - can be extended)
	proposedProps, ok := proposedSchema.Payload["properties"].(map[string]interface{})
	if !ok {
		return nil // No properties to conflict
	}

	currentProps, ok := currentSet.Effective["properties"].(map[string]interface{})
	if !ok {
		return nil // No existing properties
	}

	// Warn about field redefinitions (but don't block - shadowing is allowed)
	conflicts := []string{}
	for fieldName := range proposedProps {
		if _, exists := currentProps[fieldName]; exists {
			conflicts = append(conflicts, fieldName)
		}
	}

	// For now, just log conflicts - they're allowed via shadowing
	// Future: could enforce stricter compatibility rules
	if len(conflicts) > 0 {
		// This is informational, not an error
		// return fmt.Errorf("schema would redefine existing fields: %v", conflicts)
	}

	return nil
}
