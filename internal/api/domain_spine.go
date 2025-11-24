package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/core/domain/spineconfig"

	"github.com/jackc/pgx/v5"
)

// SpineEntryResponse describes a single entry in the canonical spine API response
type SpineEntryResponse struct {
	Name      string  `json:"name"`
	ID        *string `json:"id,omitempty"`
	ParentID  *string `json:"parent_id,omitempty"`
	Dimension int     `json:"dimension"`
	PSeat     struct {
		Occupied bool `json:"occupied"`
	} `json:"pseat"`
}

// GET /api/domain/spine
// Returns the canonical spine entries along with database ids (when present), parent ids,
// numeric dimension and a simple pseat stub.
func (s *Server) handleDomainSpine(w http.ResponseWriter, r *http.Request) {
	db := s.requireDB(w)
	if db == nil {
		return
	}
	ctx := r.Context()

	resp := []SpineEntryResponse{}
	for _, e := range spineconfig.CanonicalSpine() {
		var id *string
		var parentID *string

		// Try both legacy prefixed name and canonical bare name
		row := db.QueryRow(ctx, `
            SELECT id::text, parent_id::text
            FROM domains
            WHERE name = $1 OR name = $2
            LIMIT 1
        `, "domain."+e.Name, e.Name)

		var iid string
		var pid *string
		if err := row.Scan(&iid, &pid); err != nil {
			if err == pgx.ErrNoRows {
				id = nil
				parentID = nil
			} else {
				http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			id = &iid
			parentID = pid
		}

		se := SpineEntryResponse{
			Name:      e.Name,
			ID:        id,
			ParentID:  parentID,
			Dimension: e.Dimension,
		}
		se.PSeat.Occupied = false
		resp = append(resp, se)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "encode error: "+err.Error(), http.StatusInternalServerError)
	}
}
