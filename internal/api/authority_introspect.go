package api

import (
	"net/http"

	"dis-core/internal/core/authority"
)

// HandleAuthorityIntrospectHandler returns an http.HandlerFunc bound to an authority engine.
func HandleAuthorityIntrospectHandler(engine *authority.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		info, err := engine.GetIntrospect(ctx)
		if err != nil {
			http.Error(w, "failed to introspect authority engine: "+err.Error(), http.StatusInternalServerError)
			return
		}
		JSON(w, http.StatusOK, info)
	}
}
