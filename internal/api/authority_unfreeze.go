package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/core/authority"
)

// NewUnfreezeHandler returns a handler that clears a domain freeze via the engine.
func NewUnfreezeHandler(engine *authority.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		domainID := r.URL.Query().Get("domain_id")
		if domainID == "" {
			var body struct {
				DomainID string `json:"domain_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				domainID = body.DomainID
			}
		}
		if domainID == "" {
			http.Error(w, "missing domain_id", http.StatusBadRequest)
			return
		}

		resp, err := engine.Unfreeze(domainID)
		if err != nil {
			http.Error(w, "failed to unfreeze domain: "+err.Error(), http.StatusInternalServerError)
			return
		}
		JSON(w, http.StatusOK, resp)
	}
}
