package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/core/authority"
)

// NewOverrideHandler returns a handler that applies an administrative override.
func NewOverrideHandler(engine *authority.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Expect domain_id and changes in the body
		domainID, _ := body["domain_id"].(string)
		if domainID == "" {
			http.Error(w, "missing domain_id in body", http.StatusBadRequest)
			return
		}

		// Changes payload may be provided under `changes` key or the whole body is changes
		var changes map[string]interface{}
		if c, ok := body["changes"].(map[string]interface{}); ok {
			changes = c
		} else {
			// Remove domain_id and treat the rest as changes
			changes = make(map[string]interface{})
			for k, v := range body {
				if k == "domain_id" {
					continue
				}
				changes[k] = v
			}
		}

		resp, err := engine.Override(domainID, changes)
		if err != nil {
			http.Error(w, "failed to apply override: "+err.Error(), http.StatusInternalServerError)
			return
		}

		JSON(w, http.StatusOK, resp)
	}
}
