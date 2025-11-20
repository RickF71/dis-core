package server

import (
	"encoding/json"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"
)

var startTime = time.Now()

// handleHealth responds with runtime and system status.
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{
		"core":     "DIS-Core",
		"version":  "v1.0-dev",
		"status":   "ok",
		"uptime":   time.Since(startTime).Round(time.Second).String(),
		"domains":  0,
		"receipts": 0,
	}

	// Check database connection (nil-safe)
	if s.DB == nil {
		status["db"] = "unavailable"
		status["latency_ms"] = -1
		status["status"] = "red"
	} else {
		startPing := time.Now()
		if err := s.DB.Ping(); err != nil {
			status["db"] = "unreachable"
			status["latency_ms"] = -1
			status["status"] = "red"
		} else {
			status["db"] = "ok"
			status["latency_ms"] = time.Since(startPing).Milliseconds()
		}
	}

	addRuntime(status)
	json.NewEncoder(w).Encode(status)
}

func addRuntime(status map[string]any) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	goVersion := "unknown"
	if bi, ok := debug.ReadBuildInfo(); ok && bi != nil && bi.GoVersion != "" {
		goVersion = bi.GoVersion
	}

	status["runtime"] = map[string]any{
		"go_version": goVersion,
		"goroutines": runtime.NumGoroutine(),
		"alloc_mb":   float64(m.Alloc) / 1024.0 / 1024.0,
	}
}
