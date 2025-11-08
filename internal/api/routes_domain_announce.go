package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// ------------------------------------------------------------
// Domain Announcement Endpoint
// ------------------------------------------------------------
//
// GET /api/domain/{id}/announce
// Returns a lightweight public descriptor of the domain.
// Contains only safe, discoverable fields.
// Future versions may include public actions or verified metadata.
//
// Example response:
// {
//   "id": "52e6c580-da99-4acd-b456-6dc8e760e5f7",
//   "name": "domain.user.rick",
//   "description": "Rick's personal domain",
//   "visibility": "public"
// }
// ------------------------------------------------------------

func (s *Server) handleDomainAnnounce(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var (
		name, description sql.NullString
	)

	q := `SELECT name, COALESCE(data->>'description','') FROM domains WHERE id = $1`
	ctx := r.Context()
	err := s.DB().QueryRow(ctx, q, id).Scan(&name, &description)
	if err == sql.ErrNoRows {
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"id":          id,
		"name":        name.String,
		"description": description.String,
		"visibility":  "public",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
