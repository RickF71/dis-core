package authorityapi

import (
	"encoding/json"
	"net/http"
	"time"

	"dis-core/internal/authority"
	"dis-core/internal/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler struct groups shared resources.
type Handler struct {
	DB      *pgxpool.Pool
	Schema  *authority.AuthoritySchema
	Console *authority.Console
}

// NewHandler initializes the Authority API layer.
func NewHandler(pool *pgxpool.Pool, schema *authority.AuthoritySchema, console *authority.Console) *Handler {
	return &Handler{DB: pool, Schema: schema, Console: console}
}

// GET /api/authority/schema
func (h *Handler) HandleSchema(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.Schema == nil {
		http.Error(w, "authority schema not loaded", http.StatusInternalServerError)
		return
	}

	data, err := json.MarshalIndent(h.Schema, "", "  ")
	if err != nil {
		http.Error(w, "failed to encode schema: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// GET /api/authority/console
func (h *Handler) HandleConsole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Load latest receipts (limit 10)
	receipts, err := db.ListReceipts(ctx, h.DB, 10, 0)
	if err != nil {
		http.Error(w, "failed to load receipts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Load current verification state (if available)
	var lastVerify time.Time
	if h.Console != nil {
		lastVerify = h.Console.LastVerification
	}

	// Optional: query freeze state once implemented
	freezeActive := false // placeholder

	resp := map[string]any{
		"schema_version":    h.Schema.Version,
		"domain":            h.Console.Domain,
		"core":              h.Console.CoreVersion,
		"last_verification": lastVerify,
		"freeze_active":     freezeActive,
		"recent_receipts":   receipts,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
