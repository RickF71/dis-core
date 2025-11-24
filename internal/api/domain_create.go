package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"dis-core/internal/core/domain"
)

// DomainInput is what the client sends in JSON
type DomainInput struct {
	Name     string          `json:"name"`
	ParentID *uuid.UUID      `json:"parent_id,omitempty"`
	Data     json.RawMessage `json:"data"` // Will be stored in payload column
}

// POST /api/domain
func (s *Server) handleCreateDomain(w http.ResponseWriter, r *http.Request) {
	db := s.requireDB(w)
	if db == nil {
		return
	}
	ctx := r.Context()
	var input DomainInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate parent if provided
	if input.ParentID != nil {
		var exists bool
		err := db.QueryRow(ctx, `SELECT true FROM domains WHERE id = $1`, input.ParentID.String()).Scan(&exists)
		if err == pgx.ErrNoRows {
			http.Error(w, "parent not found", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// AI domain placement enforcement: names containing ".ai." must be
	// created under a user domain which itself is under a corporeal domain.
	if strings.Contains(input.Name, ".ai.") {
		if input.ParentID == nil {
			http.Error(w, "ai domains must be children of a user domain", http.StatusBadRequest)
			return
		}
		if err := domain.ValidateAIDomainPlacementPool(ctx, db, input.ParentID.String()); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Generate a new UUID for the domain
	newID := uuid.New()

	var parentIDStr *string
	if input.ParentID != nil {
		s := input.ParentID.String()
		parentIDStr = &s
	}

	// Insert domain with name and payload (data is now stored in payload column)
	_, err := db.Exec(ctx, `
		INSERT INTO domains (id, name, parent_id, payload)
		VALUES ($1, $2, $3, $4)
	`, newID.String(), input.Name, parentIDStr, input.Data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Record domain creation using ci.call.v1
	if s.ledger != nil {
		_ = s.ledger.RecordCall(ctx, "api", newID.String(), "domain", "domain.create.v1", map[string]any{
			"domain_id":  newID.String(),
			"parent_id":  parentIDStr,
			"created_by": "api.domain.create",
			"data_size":  len(input.Data),
			"name":       input.Name,
		})
	}

	// Fetch the created domain
	var d Domain
	err = db.QueryRow(ctx, `
		SELECT id::text, parent_id::text, payload, created_at, updated_at
		FROM domains
		WHERE id = $1::uuid
	`, newID.String()).Scan(&d.ID, &d.ParentID, &d.Payload, &d.CreatedAt, &d.UpdatedAt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}
