package api

import (
	"encoding/json"
	"net/http"

	coreauth "dis-core/internal/core/authority"
)

// NewSeatInstantiateHandler returns a handler that delegates seat instantiation
// to the core authority engine.
func NewSeatInstantiateHandler(engine *coreauth.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if engine == nil {
			http.Error(w, "authority engine not available", http.StatusServiceUnavailable)
			return
		}
		var req struct {
			DomainID   string `json:"domain_id"`
			IdentityID string `json:"identity_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		seat, err := engine.InstantiateSeat(r.Context(), req.DomainID, req.IdentityID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(seat)
	}
}
