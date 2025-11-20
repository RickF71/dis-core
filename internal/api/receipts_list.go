package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"dis-core/internal/phase9c"
)

// handleListReceipts handles GET /api/receipts/list for Phase 9D receipt listing
func (s *Server) handleListReceipts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Require DB for listing receipts
	db := s.requireDB(w)
	if db == nil {
		return
	}

	receipts, err := phase9c.ListReceipts(ctx, db, limit, offset)
	if err != nil {
		http.Error(w, "Failed to list receipts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"receipts": receipts,
		"limit":    limit,
		"offset":   offset,
		"count":    len(receipts),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleReceiptsDashboard handles GET /api/receipts/dashboard for Phase 9D continuity dashboard metrics
func (s *Server) handleReceiptsDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Require DB for dashboard stats
	db := s.requireDB(w)
	if db == nil {
		return
	}

	stats, err := phase9c.GetReceiptStats(ctx, db)
	if err != nil {
		http.Error(w, "Failed to get receipt stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
