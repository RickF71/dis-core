package api

import (
	"context"
	"net/http"
)

// registerReceiptRoutes registers all /api/receipts endpoints.
func (s *Server) registerReceiptRoutes() {
	mux := s.router
	mux.HandleFunc("GET /api/receipts/latest", s.handleGetLatestReceipt)
	mux.HandleFunc("GET /api/receipts/verify/{id}", s.handleVerifyReceipt)
}

// GET /api/receipts/latest
func (s *Server) handleGetLatestReceipt(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if s.ledger == nil {
		http.Error(w, "ledger not initialized", http.StatusInternalServerError)
		return
	}

	row := s.db.QueryRow(ctx, `
		SELECT payload
		FROM receipts
		ORDER BY created_at DESC
		LIMIT 1;
	`)

	var payloadBytes []byte
	if err := row.Scan(&payloadBytes); err != nil {
		http.Error(w, "no receipts found or query failed: "+err.Error(), http.StatusNotFound)
		return
	}

	// The payload is already JSON; re-emit it directly.
	w.Header().Set("Content-Type", "application/json")
	w.Write(payloadBytes)
}
