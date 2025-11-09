package identity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// copilot: Implement identity binding and management functions.
// Add CreateIdentityBinding(db, binding), GetIdentityBindings(db, identityID), and UpdateIdentityBinding(db, bindingID, updates).
// Each record corresponds to identity_bindings table.

// IdentityBinding represents a binding between an identity and a domain or resource
type IdentityBinding struct {
	ID           string                 `json:"id" db:"id"`
	IdentityID   string                 `json:"identity_id" db:"identity_id"`
	ResourceType string                 `json:"resource_type" db:"resource_type"` // "domain", "policy", "schema"
	ResourceID   string                 `json:"resource_id" db:"resource_id"`
	BindingType  string                 `json:"binding_type" db:"binding_type"` // "owner", "admin", "viewer", "contributor"
	Permissions  map[string]interface{} `json:"permissions" db:"permissions"`
	Active       bool                   `json:"active" db:"active"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy    string                 `json:"created_by" db:"created_by"`
	ExpiresAt    *time.Time             `json:"expires_at" db:"expires_at"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
}

// CreateIdentityBinding creates a new identity binding in the database
func CreateIdentityBinding(ctx context.Context, db *pgxpool.Pool, binding *IdentityBinding) error {
	// Marshal permissions and metadata to JSON
	permissionsBytes, err := json.Marshal(binding.Permissions)
	if err != nil {
		return err
	}

	metadataBytes, err := json.Marshal(binding.Metadata)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO identity_bindings (
			id, identity_id, resource_type, resource_id, binding_type,
			permissions, active, created_at, updated_at, created_by, expires_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err = db.Exec(ctx, query,
		binding.ID,
		binding.IdentityID,
		binding.ResourceType,
		binding.ResourceID,
		binding.BindingType,
		permissionsBytes,
		binding.Active,
		time.Now(),
		time.Now(),
		binding.CreatedBy,
		binding.ExpiresAt,
		metadataBytes,
	)

	return err
}

// GetIdentityBindings retrieves all bindings for a specific identity
func GetIdentityBindings(ctx context.Context, db *pgxpool.Pool, identityID string) ([]IdentityBinding, error) {
	query := `
		SELECT id, identity_id, resource_type, resource_id, binding_type,
			   permissions, active, created_at, updated_at, created_by, expires_at, metadata
		FROM identity_bindings
		WHERE identity_id = $1
		ORDER BY created_at DESC`

	rows, err := db.Query(ctx, query, identityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bindings []IdentityBinding
	for rows.Next() {
		var binding IdentityBinding
		var permissionsBytes, metadataBytes []byte

		err := rows.Scan(
			&binding.ID,
			&binding.IdentityID,
			&binding.ResourceType,
			&binding.ResourceID,
			&binding.BindingType,
			&permissionsBytes,
			&binding.Active,
			&binding.CreatedAt,
			&binding.UpdatedAt,
			&binding.CreatedBy,
			&binding.ExpiresAt,
			&metadataBytes,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if len(permissionsBytes) > 0 {
			err = json.Unmarshal(permissionsBytes, &binding.Permissions)
			if err != nil {
				return nil, err
			}
		}

		if len(metadataBytes) > 0 {
			err = json.Unmarshal(metadataBytes, &binding.Metadata)
			if err != nil {
				return nil, err
			}
		}

		bindings = append(bindings, binding)
	}

	return bindings, nil
}

// UpdateIdentityBinding updates an existing identity binding
func UpdateIdentityBinding(ctx context.Context, db *pgxpool.Pool, bindingID string, updates map[string]interface{}) error {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{bindingID}
	argIndex := 2

	if resourceType, ok := updates["resource_type"]; ok {
		setParts = append(setParts, "resource_type = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, resourceType)
		argIndex++
	}

	if resourceID, ok := updates["resource_id"]; ok {
		setParts = append(setParts, "resource_id = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, resourceID)
		argIndex++
	}

	if bindingType, ok := updates["binding_type"]; ok {
		setParts = append(setParts, "binding_type = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, bindingType)
		argIndex++
	}

	if permissions, ok := updates["permissions"]; ok {
		permissionsBytes, err := json.Marshal(permissions)
		if err != nil {
			return err
		}
		setParts = append(setParts, "permissions = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, permissionsBytes)
		argIndex++
	}

	if active, ok := updates["active"]; ok {
		setParts = append(setParts, "active = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, active)
		argIndex++
	}

	if expiresAt, ok := updates["expires_at"]; ok {
		setParts = append(setParts, "expires_at = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, expiresAt)
		argIndex++
	}

	if metadata, ok := updates["metadata"]; ok {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		setParts = append(setParts, "metadata = $"+fmt.Sprintf("%d", argIndex))
		args = append(args, metadataBytes)
		argIndex++
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	// Add updated_at
	setParts = append(setParts, "updated_at = $"+fmt.Sprintf("%d", argIndex))
	args = append(args, time.Now())

	query := fmt.Sprintf("UPDATE identity_bindings SET %s WHERE id = $1", strings.Join(setParts, ", "))

	_, err := db.Exec(ctx, query, args...)
	return err
}

// HandleCreateIdentityBinding handles POST /api/identity/bindings
func HandleCreateIdentityBinding(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	var binding IdentityBinding

	if err := json.NewDecoder(r.Body).Decode(&binding); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if binding.ID == "" {
		binding.ID = generateBindingID()
	}

	// Set defaults
	if binding.Permissions == nil {
		binding.Permissions = make(map[string]interface{})
	}
	if binding.Metadata == nil {
		binding.Metadata = make(map[string]interface{})
	}
	binding.Active = true

	err := CreateIdentityBinding(ctx, db, &binding)
	if err != nil {
		http.Error(w, "failed to create binding: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(binding)
}

// generateBindingID generates a unique binding ID
func generateBindingID() string {
	return fmt.Sprintf("bind_%d", time.Now().UnixNano())
}
