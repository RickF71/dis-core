package api

import (
	"encoding/json"
	"net/http"

	coreauth "dis-core/internal/core/authority"

	"github.com/go-chi/chi/v5"
)

func NewSeatListHandler(engine *coreauth.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if engine == nil {
			http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
			return
		}
		domainID := chi.URLParam(r, "id")
		if domainID == "" {
			http.Error(w, "domain id required", http.StatusBadRequest)
			return
		}
		seats, err := engine.ListDomainSeats(r.Context(), domainID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(seats)
	}
}
