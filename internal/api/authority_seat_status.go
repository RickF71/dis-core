package api

import (
	"encoding/json"
	"net/http"

	coreauth "dis-core/internal/core/authority"

	"github.com/go-chi/chi/v5"
)

func NewSeatStatusHandler(engine *coreauth.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if engine == nil {
			http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
			return
		}
		seatID := chi.URLParam(r, "id")
		if seatID == "" {
			http.Error(w, "seat id required", http.StatusBadRequest)
			return
		}
		st, err := engine.SeatStatus(r.Context(), seatID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(st)
	}
}
