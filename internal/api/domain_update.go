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

	ctx := r.Context()
	id := r.PathValue("id")
	var body struct {
		CSS string `json:"css"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Update CSS in the standardized nested path {meta,data,css}
	_, err := s.DB().Exec(ctx, `
		UPDATE domains
		SET data = jsonb_set(
			jsonb_set(
				jsonb_set(
					COALESCE(data, '{}'::jsonb),
					'{meta}',
					COALESCE(data->'meta', '{}'::jsonb),
					true
				),
				'{meta,data}',
				COALESCE(data#>'{meta,data}', '{}'::jsonb),
				true
			),
			'{meta,data,css}',
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

	// Record domain update using ci.call.v1
	if s.ledger != nil {
		_ = s.ledger.RecordCall(ctx, "api", id, "domain", "domain.update.v1", map[string]any{
			"domain_id":  id,
			"updated_by": "api.domain.update",
			"field":      "css",
			"css_length": len(body.CSS),
		})
	}

	w.WriteHeader(http.StatusNoContent)
}
