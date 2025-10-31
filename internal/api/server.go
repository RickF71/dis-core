package api

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"runtime"
	"runtime/debug"
	"time"

	"dis-core/internal/config"
	"dis-core/internal/domain"
	"dis-core/internal/ledger"
	"dis-core/internal/overlay"
	"dis-core/internal/policy"
	"dis-core/internal/schema"
	//"dis-core/internal/schema"
)

// WithLedger sets the Ledger pointer and returns the server (chainable)
func (s *Server) WithLedger(led *ledger.Ledger) *Server {
	s.Ledger = led
	return s
}

type Server struct {
	PolicyEngine policy.PolicyEngine
	mux          *http.ServeMux
	db           *sql.DB
	cfg          *config.Config

	// Core components
	Ledger *ledger.Ledger

	// Managers (YAML import & domain logic)
	DomainManager  *domain.Manager
	SchemaManager  *schema.Manager
	PolicyManager  *policy.Manager
	OverlayManager *overlay.Manager // safe to keep even if stub

	// Optional legacy store field (some older routes expect it)
	Store *ledger.Store

	// Optional logger
	logger *log.Logger

	// Optional schema registry (for validation)
	schemas *schema.Registry
}

// Mux returns the internal HTTP mux for this server.
func (s *Server) Mux() *http.ServeMux { return s.mux }

// handlePing is a simple health endpoint for API status checks.
func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"message": "DIS node alive",
	})
}

// handleInfo reports basic build and version info.
func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version": "0.9.3",
		"core":    "DIS-Core",
	})
}

// =====================================================
//  /api/health — Node health, runtime, and subsystem status
// =====================================================

// Global start timestamp for uptime calculation.
var serverStart = time.Now()

// handleHealth performs a comprehensive self-check and reports system status.
// Exposed at /api/health.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{
		"core":     "DIS-Core",
		"version":  s.cfg.Version,
		"status":   "ok",
		"health":   "green",
		"schemas":  0,
		"domains":  0,
		"receipts": 0,
	}

	// --- Schema registry health
	if s.schemas != nil {
		status["schemas"] = len(s.schemas.ByKey())
	} else {
		status["schemas"] = "unavailable"
		status["health"] = "yellow"
	}

	// --- Manager subsystem health
	status["managers"] = map[string]any{
		"domain_manager":  healthState(s.DomainManager != nil),
		"schema_manager":  healthState(s.SchemaManager != nil),
		"policy_manager":  healthState(s.PolicyManager != nil),
		"overlay_manager": healthState(s.OverlayManager != nil),
	}

	// --- Database / Ledger health
	if s.Ledger == nil || s.Ledger.DB == nil {
		status["domains"] = "unknown"
		status["receipts"] = "unknown"
		status["health"] = "red"
		addRuntimeMetrics(status)
		writeJSON(w, http.StatusOK, status)
		return
	}

	// Count registered domains
	var domainCount int
	if err := s.Ledger.DB.QueryRow(`SELECT COUNT(*) FROM canon WHERE type = 'domain'`).Scan(&domainCount); err != nil {
		status["domains"] = "error"
		status["health"] = "yellow"
	} else {
		status["domains"] = domainCount
	}

	// Count stored receipts
	var receiptCount int
	if err := s.Ledger.DB.QueryRow(`SELECT COUNT(*) FROM receipts`).Scan(&receiptCount); err != nil {
		status["receipts"] = "error"
		status["health"] = "yellow"
	} else {
		status["receipts"] = receiptCount
	}

	// --- DB ping latency
	startPing := time.Now()
	if err := s.Ledger.DB.Ping(); err != nil {
		status["db_latency_ms"] = "unreachable"
		status["health"] = "red"
	} else {
		status["db_latency_ms"] = time.Since(startPing).Milliseconds()
	}

	// --- Add runtime metrics
	addRuntimeMetrics(status)

	writeJSON(w, http.StatusOK, status)
}

// helper: convert bytes to MB
func bToMb(b uint64) float64 { return float64(b) / 1024.0 / 1024.0 }

func addRuntimeMetrics(status map[string]any) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Go version via build info (fallback to "unknown")
	gov := "unknown"
	if bi, ok := debug.ReadBuildInfo(); ok && bi != nil && bi.GoVersion != "" {
		gov = bi.GoVersion
	}

	// seconds since last GC (guard when LastGC == 0)
	var lastGCSec float64
	if m.LastGC != 0 {
		lastGCSec = time.Since(time.Unix(0, int64(m.LastGC))).Seconds()
	} else {
		lastGCSec = -1 // means "no GC yet"
	}

	status["runtime"] = map[string]any{
		"uptime":       time.Since(serverStart).Round(time.Second).String(),
		"goroutines":   runtime.NumGoroutine(),
		"alloc_mb":     bToMb(m.Alloc),
		"sys_mb":       bToMb(m.Sys),
		"heap_objects": m.HeapObjects,
		"num_gc":       m.NumGC,
		"last_gc_sec":  lastGCSec,
		"go_version":   gov,
		"num_cpu":      runtime.NumCPU(),
		"arch":         runtime.GOARCH,
		"os":           runtime.GOOS,
	}
}

func healthState(ok bool) string {
	if ok {
		return "ok"
	}
	return "missing"
}

// NewServer constructs and initializes a DIS-Core API server.
func NewServer(cfg *config.Config, led *ledger.Ledger, db *sql.DB) *Server {
	s := &Server{
		cfg:     cfg, // <— store config so s.cfg.Version works
		mux:     http.NewServeMux(),
		db:      db,
		Ledger:  led,
		schemas: nil, // will fill below if ledger has registry
	}

	// Initialize store and API routes
	s.Store = ledger.NewStore(db)
	s.RegisterAPIs() // reconnect routes from routes.go

	// Try to attach registry from the ledger if available
	if led != nil {
		if reg := led.Registry(); reg != nil {
			s.schemas = reg
		}
	}

	// Do NOT wrap s.mux with CORS here; wrap at ListenAndServe
	return s
}

// WithLogger sets a custom logger and returns the server (chainable)
func (s *Server) WithLogger(l *log.Logger) *Server {
	s.logger = l
	return s
}

// WithSchemas sets a schema registry and returns the server (chainable)
func (s *Server) WithSchemas(reg *schema.Registry) *Server {
	s.schemas = reg
	return s
}

// handleSchemaList returns all registered schema IDs and versions as JSON.
func (s *Server) handleSchemaList(w http.ResponseWriter, r *http.Request) {
	if s.schemas == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "schema registry unavailable"})
		return
	}
	type schemaInfo struct {
		ID      string `json:"id"`
		Version string `json:"version"`
	}
	var out []schemaInfo
	for _, e := range s.schemasEntries() {
		out = append(out, schemaInfo{ID: e.ID, Version: e.Version})
	}
	writeJSON(w, http.StatusOK, out)
}

// schemasEntries returns all schema entries in the registry.
func (s *Server) schemasEntries() []schema.Entry {
	if s.schemas == nil {
		return nil
	}
	entries := make([]schema.Entry, 0, len(s.schemasEntriesMap()))
	for _, e := range s.schemasEntriesMap() {
		entries = append(entries, e)
	}
	return entries
}

// schemasEntriesMap returns the byKey map from the registry (read-only).
func (s *Server) schemasEntriesMap() map[string]schema.Entry {
	if s.schemas == nil {
		return nil
	}
	return s.schemas.ByKey()
}

func (s *Server) Run(ctx context.Context) error {
	// TODO: Start HTTP server, handle graceful shutdown
	return nil
}

// TODO: Keep RegisterAPIs() as canonical, and split per-route files as needed.
