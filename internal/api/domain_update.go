package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleUpdateDomain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	var body struct {
		CSS string `json:"css"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Update only the "css" field inside JSONB column "data"
	_, err := s.DB().Exec(`
		UPDATE domains
		SET data = jsonb_set(
			COALESCE(data, '{}'::jsonb),
			'{css}',
			to_jsonb($1::text),
			true
		),
		updated_at = NOW()
		WHERE id = $2
	`, body.CSS, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
