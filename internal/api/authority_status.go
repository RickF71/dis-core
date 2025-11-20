package api

import (
	"net/http"

	"dis-core/internal/core/authority"
)

// HandleAuthorityStatusHandler returns an http.HandlerFunc bound to the provided engine.
func HandleAuthorityStatusHandler(engine *authority.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		status, err := engine.GetStatus(ctx)
		if err != nil {
			http.Error(w, "failed to fetch authority status: "+err.Error(), http.StatusInternalServerError)
			return
		}
		JSON(w, http.StatusOK, status)
	}
}
