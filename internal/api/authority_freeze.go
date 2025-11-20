package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/core/authority"
)

// NewFreezeHandler returns a handler that applies a domain freeze via the engine.
func NewFreezeHandler(engine *authority.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Accept domain id either as query param or JSON body {"domain_id":"..."}
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

		resp, err := engine.Freeze(domainID)
		if err != nil {
			http.Error(w, "failed to freeze domain: "+err.Error(), http.StatusInternalServerError)
			return
		}
		JSON(w, http.StatusOK, resp)
	}
}
