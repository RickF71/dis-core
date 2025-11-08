package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// DomainInput is what the client sends in JSON
type DomainInput struct {
	Name     string          `json:"name"`
	ParentID *uuid.UUID      `json:"parent_id,omitempty"`
	Data     json.RawMessage `json:"data"`
}

// POST /api/domain
func (s *Server) handleCreateDomain(w http.ResponseWriter, r *http.Request) {
	var input DomainInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate parent if provided
	if input.ParentID != nil {
		var exists bool
		err := s.DB().QueryRow(`SELECT true FROM domains WHERE id = $1`, *input.ParentID).Scan(&exists)
		if err == sql.ErrNoRows {
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

	_, err := s.DB().Exec(`
		INSERT INTO domains (id, parent_id, data)
		VALUES ($1, $2, $3)
	`, newID, input.ParentID, input.Data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var d Domain
	err = s.DB().QueryRow(`
		SELECT id, parent_id, data, created_at, updated_at
		FROM domains
		WHERE id = $1
	`, newID).Scan(&d.ID, &d.ParentID, &d.Data, &d.CreatedAt, &d.UpdatedAt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(d)
}
