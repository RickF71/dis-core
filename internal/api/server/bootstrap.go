package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// BootstrapStatus reports counts of key system tables and overall health.
type BootstrapStatus struct {
	Status       string `json:"status"`
	Domains      int    `json:"domains"`
	Schemas      int    `json:"schemas"`
	Receipts     int    `json:"receipts"`
	Policies     int    `json:"policies"`
	Bootstrapped bool   `json:"bootstrapped"`
	DBLatency    int64  `json:"db_latency_ms"`
	Timestamp    string `json:"timestamp"`
}

type DBProvider interface {
	QueryRow(string, ...any) RowScanner
	Ping(context.Context) error
}
type RowScanner interface {
	Scan(dest ...any) error
}

// HandleBootstrapStatus responds with DB stats and readiness check.
func HandleBootstrapStatus(db DBProvider, w http.ResponseWriter, r *http.Request) {
	resp := BootstrapStatus{
		Status:       "ok",
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Bootstrapped: true,
	}

	if db == nil {
		resp.Status = "unavailable"
		resp.Bootstrapped = false
		writeJSON(w, resp)
		return
	}

	start := time.Now()
	if err := db.Ping(context.Background()); err != nil {
		resp.Status = "error"
		resp.Bootstrapped = false
		resp.DBLatency = -1
		writeJSON(w, resp)
		return
	}
	resp.DBLatency = time.Since(start).Milliseconds()

	// Count key tables — safe even if empty
	count := func(q string) int {
		var c int
		if err := db.QueryRow(q).Scan(&c); err == nil {
			return c
		}
		return -1
	}

	resp.Domains = count(`SELECT COUNT(*) FROM canon WHERE type='domain'`)
	resp.Schemas = count(`SELECT COUNT(*) FROM schemas`)
	resp.Receipts = count(`SELECT COUNT(*) FROM receipts`)
	resp.Policies = count(`SELECT COUNT(*) FROM policies`)

	// simple heuristic: must have at least 1 domain and 1 schema
	if resp.Domains <= 0 || resp.Schemas <= 0 {
		resp.Bootstrapped = false
		resp.Status = "incomplete"
	}

	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}
