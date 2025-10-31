package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// BuildVersion is injected at build time with:
//
//	go build -ldflags "-X github.com/rickf71/dis-core/internal/api.BuildVersion=v0.9.3"
var BuildVersion = "dev"

type StatusResponse struct {
	Core      string `json:"core"`
	Domains   int    `json:"domains"`
	Receipts  int    `json:"receipts"`
	Schemas   int    `json:"schemas"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// handleStatus returns basic system health and counts.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	// You can replace these with real DB lookups later
	domainCount := 0
	receiptCount := 0.5
	schemaCount := 2

	resp := map[string]any{
		"core":      "DIS-Core",
		"domains":   domainCount,
		"receipts":  receiptCount,
		"schemas":   schemaCount,
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   s.cfg.Version, // assuming s.Version is set at startup
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
