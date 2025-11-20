package api

import (
	"encoding/json"
	"log"
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
	Payload   json.RawMessage `json:"payload"` // Phase 10J.4: renamed from Data
	CreatedAt *time.Time      `json:"created_at"`
	UpdatedAt *time.Time      `json:"updated_at"`
}

// GET /api/domain/{id}
func (s *Server) handleGetDomain(w http.ResponseWriter, r *http.Request) {
	db := s.requireDB(w)
	if db == nil {
		return
	}
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
		log.Printf("[domain_get] Querying by UUID: %s", idParam)
		err = db.QueryRow(ctx, `
			SELECT id::text, parent_id::text, payload, created_at, updated_at
			FROM domains
			WHERE id = $1::uuid
		`, idParam).Scan(&d.ID, &d.ParentID, &d.Payload, &d.CreatedAt, &d.UpdatedAt)
		log.Printf("[domain_get] Query result - err: %v", err)
	} else {
		// Treat it as canonical domain ID (e.g. "domain.null")
		log.Printf("[domain_get] Querying by name/id string: %s", idParam)
		err = db.QueryRow(ctx, `
			SELECT id::text, parent_id::text, payload, created_at, updated_at
			FROM domains
			WHERE id = $1
		`, idParam).Scan(&d.ID, &d.ParentID, &d.Payload, &d.CreatedAt, &d.UpdatedAt)
	}

	switch {
	case err == pgx.ErrNoRows:
		log.Printf("[domain_get] No rows found for: %s", idParam)
		JSONError(w, http.StatusNotFound, "domain not found")
		return
	case err != nil:
		log.Printf("[domain_get] Query error: %v", err)
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[domain_get] Successfully loaded domain: %s", d.ID)
	JSON(w, http.StatusOK, d)
}
