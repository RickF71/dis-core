package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// DomainInput is what the client sends in JSON
type DomainInput struct {
	Name     string          `json:"name"`
	ParentID *uuid.UUID      `json:"parent_id,omitempty"`
	Data     json.RawMessage `json:"data"`
}

// POST /api/domain
func (s *Server) handleCreateDomain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var input DomainInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate parent if provided
	if input.ParentID != nil {
		var exists bool
		err := s.DB().QueryRow(ctx, `SELECT true FROM domains WHERE id = $1`, input.ParentID.String()).Scan(&exists)
		if err == pgx.ErrNoRows {
			http.Error(w, "parent not found", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

	_, err := s.DB().Exec(ctx, `
		INSERT INTO domains (id, parent_id, data)
		VALUES ($1, $2, $3)
	`, newID.String(), parentIDStr, input.Data)

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

	var d Domain
	err = s.DB().QueryRow(ctx, `
		SELECT id, parent_id, data, created_at, updated_at
		FROM domains
		WHERE id = $1
	`, newID.String()).Scan(&d.ID, &d.ParentID, &d.Data, &d.CreatedAt, &d.UpdatedAt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}
