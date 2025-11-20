package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"dis-core/internal/receipts"
)

// handleVerifyReceipt handles GET /api/receipts/verify/{id} for Phase 9C receipt verification
func (s *Server) handleVerifyReceipt(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, "Receipt ID required", http.StatusBadRequest)
		return
	}

	db := s.requireDB(w)
	if db == nil {
		return
	}

	result, err := receipts.VerifyReceipt(ctx, db, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
