package api

import (
	"dis-core/internal/reconcile"
	"fmt"
	"log"
	"net/http"
)

// registerReconcileRoutes wires the reconciliation endpoints into the mux.
func (s *Server) registerReconcileRoutes() {
	mux := s.mux

	mux.HandleFunc("/api/reconcile/schemas", s.handleReconcileSchemas)
	mux.HandleFunc("/api/reconcile/domains", s.handleReconcileDomains)
	mux.HandleFunc("/api/reconcile/apply", s.handleReconcileApply)

	log.Println("✅ Reconcile routes registered")
}

// ------------------------------------------------------------
//  HANDLERS
// ------------------------------------------------------------

// handleReconcileSchemas lists schema import warnings such as version mismatches or bad YAML.
func (s *Server) handleReconcileSchemas(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(`
        SELECT id, file_path, domain, schema_type, details, resolved, created_at, resolved_at
        FROM import_warnings
        WHERE type IN ('version_mismatch', 'invalid_yaml')
        ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []SchemaWarning
	for rows.Next() {
		var rec SchemaWarning
		rows.Scan(&rec.ID, &rec.FilePath, &rec.Domain, &rec.SchemaType,
			&rec.Details, &rec.Resolved, &rec.CreatedAt, &rec.ResolvedAt)
		results = append(results, rec)
	}
	JSON(w, http.StatusOK, map[string]any{"count": len(results), "items": results})
}

// handleReconcileDomains lists domain linkage or ancestry problems.
func (s *Server) handleReconcileDomains(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(`
        SELECT id, domain, schema_type AS parent, details, resolved, created_at, resolved_at
        FROM import_warnings
        WHERE type IN ('missing_parent', 'missing_schema_reference')
        ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []DomainWarning
	for rows.Next() {
		var rec DomainWarning
		rows.Scan(&rec.ID, &rec.Domain, &rec.Parent, &rec.Details,
			&rec.Resolved, &rec.CreatedAt, &rec.ResolvedAt)
		results = append(results, rec)
	}
	JSON(w, http.StatusOK, map[string]any{"count": len(results), "items": results})
}

// handleReconcileApply performs canonical reconciliation for a specific domain.
// Example: POST /api/reconcile/apply?domain=usa
func (s *Server) handleReconcileApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	domainCode := r.URL.Query().Get("domain")
	if domainCode == "" {
		http.Error(w, "missing ?domain= parameter", http.StatusBadRequest)
		return
	}

	reconciler := reconcile.New(
		s.cfg.RepoRoot, // YAML repo root
		s.schemas,      // active schema registry
		s.Ledger,       // live ledger
	)

	if err := reconciler.ReconcileDomain(domainCode); err != nil {
		http.Error(w, fmt.Sprintf("reconcile failed: %v", err), http.StatusBadRequest)
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"domain":  domainCode,
		"message": fmt.Sprintf("domain %s reconciled and logged in ledger", domainCode),
	})
}
