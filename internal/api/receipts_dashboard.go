package api

import (
	"net/http"
	"path/filepath"
)

// handleReceiptsDashboardHTML serves the HTML dashboard for Phase 9D
func (s *Server) handleReceiptsDashboardHTML(w http.ResponseWriter, r *http.Request) {
	// Serve the static HTML file
	htmlPath := filepath.Join("static", "receipts_dashboard.html")
	http.ServeFile(w, r, htmlPath)
}
