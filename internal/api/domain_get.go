package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Domain represents a single domain record.
type Domain struct {
	ID        string          `json:"id"` // switched from uuid.UUID
	ParentID  *string         `json:"parent_id,omitempty"`
	Data      json.RawMessage `json:"data"`
	CreatedAt *time.Time      `json:"created_at"`
	UpdatedAt *time.Time      `json:"updated_at"`
}

// GET /api/domain/{id}
func (s *Server) handleGetDomain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idParam := r.PathValue("id")
	idParam = strings.TrimSpace(idParam)

	// Handle missing or invalid IDs early
	if idParam == "" || idParam == "null" || idParam == "undefined" {
		JSONError(w, http.StatusBadRequest, "missing domain id")
		return
	}

	// Support both UUIDs and string-based domain IDs
	var d Domain
	var err error

	// Try as UUID if possible
	if _, parseErr := uuid.Parse(idParam); parseErr == nil {
		err = s.DB().QueryRow(ctx, `
			SELECT id::text, parent_id::text, data, created_at, updated_at
			FROM domains
			WHERE id = $1::uuid
		`, idParam).Scan(&d.ID, &d.ParentID, &d.Data, &d.CreatedAt, &d.UpdatedAt)
	} else {
		// Treat it as canonical domain ID (e.g. "domain.null")
		err = s.DB().QueryRow(ctx, `
			SELECT id, parent_id, data, created_at, updated_at
			FROM domains
			WHERE id = $1
		`, idParam).Scan(&d.ID, &d.ParentID, &d.Data, &d.CreatedAt, &d.UpdatedAt)
	}

	switch {
	case err == pgx.ErrNoRows:
		JSONError(w, http.StatusNotFound, "domain not found")
		return
	case err != nil:
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, d)
}
