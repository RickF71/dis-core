package api

import (
	"net/http"

	"dis-core/internal/core/authority"
)

// HandleAuthorityLineageHandler returns an http.HandlerFunc wired to the given engine.
func HandleAuthorityLineageHandler(engine *authority.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		targetID := r.URL.Query().Get("id")
		if targetID == "" {
			http.Error(w, "missing id parameter", http.StatusBadRequest)
			return
		}

		lineage, err := engine.GetLineage(ctx, targetID)
		if err != nil {
			http.Error(w, "failed to fetch lineage: "+err.Error(), http.StatusInternalServerError)
			return
		}

		JSON(w, http.StatusOK, lineage)
	}
}
