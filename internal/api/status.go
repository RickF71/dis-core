package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

// BuildVersion is optionally injected at build time:
// go build -ldflags "-X dis-core/internal/api.BuildVersion=v1.0.0"
var BuildVersion = "dev"

// StatusResponse reports core system health and simple runtime stats.
type StatusResponse struct {
	Core        string `json:"core"`
	Domains     int    `json:"domains"`
	Receipts    int    `json:"receipts"`
	Schemas     int    `json:"schemas"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
	Version     string `json:"version"`
	DBLatency   int64  `json:"db_latency_ms"`
	Environment string `json:"environment,omitempty"`
}

// HandleStatus reports current DIS-Core health and counts.
func (s *Server) HandleStatus(w http.ResponseWriter, r *http.Request) {
	resp := StatusResponse{
		Core:        "DIS-Core",
		Status:      "ok",
		Version:     BuildVersion,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Environment: os.Getenv("DIS_ENV"),
	}

	// --- Safe DB queries ---
	if s.DB != nil {
		if err := s.DB().QueryRow(`SELECT COUNT(*) FROM canon WHERE type='domain'`).Scan(&resp.Domains); err != nil && err != sql.ErrNoRows {
			resp.Status = "warn"
		}
		if err := s.DB().QueryRow(`SELECT COUNT(*) FROM receipts`).Scan(&resp.Receipts); err != nil && err != sql.ErrNoRows {
			resp.Status = "warn"
		}
		if err := s.DB().QueryRow(`SELECT COUNT(*) FROM schemas`).Scan(&resp.Schemas); err != nil && err != sql.ErrNoRows {
			resp.Status = "warn"
		}

		startPing := time.Now()
		if err := s.DB().Ping(); err == nil {
			resp.DBLatency = time.Since(startPing).Milliseconds()
		} else {
			resp.DBLatency = -1
			resp.Status = "error"
		}
	} else {
		resp.Status = "unavailable"
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
