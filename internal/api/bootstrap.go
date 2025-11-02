package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//
// ============================================================
//  BOOTSTRAP API ROUTES
// ============================================================
//  These routes manage the editable “bootstrap” layer of DIS.
//  All YAMLs here are potential canon, but remain mutable.
// ============================================================
//

// RegisterBootstrapRoutes mounts bootstrap endpoints under /api/bootstrap.
func (s *Server) RegisterBootstrapRoutes(db *sql.DB) {
	mux := s.Mux()

	// --- Convenience redirect
	mux.HandleFunc("/api/bootstrap", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/bootstrap/list", http.StatusTemporaryRedirect)
	})

	// --- Core endpoints
	mux.HandleFunc("/api/bootstrap/list", withGET(func(w http.ResponseWriter, r *http.Request) {
		listBootstrapEntries(w, db)
	}))
	mux.HandleFunc("/api/bootstrap/stats", withGET(func(w http.ResponseWriter, r *http.Request) {
		getBootstrapStats(w, db)
	}))
	mux.HandleFunc("/api/bootstrap/search", withGET(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		searchBootstrapEntries(w, db, query)
	}))
	mux.HandleFunc("/api/bootstrap/reimport", withPOST(func(w http.ResponseWriter, r *http.Request) {
		reimportBootstrap(w, db)
	}))

	// --- Reconcile (save YAML from Finagler)
	mux.HandleFunc("/api/bootstrap/reconcile", withPOST(func(w http.ResponseWriter, r *http.Request) {
		s.handleBootstrapReconcile(w, r)
	}))

	// --- Catch-all (GET / DELETE /api/bootstrap/:id)
	mux.HandleFunc("/api/bootstrap/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/bootstrap/")
		if id == "" {
			Error(w, http.StatusBadRequest, fmt.Errorf("missing id"))
			return
		}

		switch r.Method {
		case http.MethodGet:
			s.handleBootstrapGet(w, id, db)
		case http.MethodDelete:
			deleteBootstrapEntry(w, id, db)
		default:
			Error(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		}
	})
}

//
// ============================================================
//  HELPERS
// ============================================================
//

// Simple method guards
func withGET(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			Error(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
			return
		}
		fn(w, r)
	}
}
func withPOST(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			Error(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
			return
		}
		fn(w, r)
	}
}

//
// ============================================================
//  QUERY ENDPOINTS (list/search/stats)
// ============================================================
//

func listBootstrapEntries(w http.ResponseWriter, db *sql.DB) {
	rows, err := db.Query(`SELECT id, type, source_file, imported_at FROM bootstrap ORDER BY imported_at DESC;`)
	if err != nil {
		Error(w, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		var id, typ, src string
		var importedAt sql.NullTime
		if err := rows.Scan(&id, &typ, &src, &importedAt); err != nil {
			Error(w, http.StatusInternalServerError, err)
			return
		}
		results = append(results, map[string]any{
			"id":          id,
			"type":        typ,
			"source_file": src,
			"imported_at": importedAt.Time,
		})
	}
	JSON(w, http.StatusOK, results)
}

func getBootstrapStats(w http.ResponseWriter, db *sql.DB) {
	rows, err := db.Query(`SELECT type, COUNT(*) FROM bootstrap GROUP BY type ORDER BY type;`)
	if err != nil {
		Error(w, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var typ string
		var count int
		if err := rows.Scan(&typ, &count); err != nil {
			Error(w, http.StatusInternalServerError, err)
			return
		}
		stats[typ] = count
	}
	JSON(w, http.StatusOK, map[string]any{"counts": stats})
}

func searchBootstrapEntries(w http.ResponseWriter, db *sql.DB, query string) {
	query = strings.TrimSpace(query)
	if query == "" {
		Error(w, http.StatusBadRequest, fmt.Errorf("missing search query ?q=term"))
		return
	}
	like := "%" + strings.ToLower(query) + "%"
	rows, err := db.Query(`
		SELECT id, type, source_file, imported_at
		FROM bootstrap
		WHERE LOWER(id) LIKE $1 OR LOWER(source_file) LIKE $1
		ORDER BY imported_at DESC;
	`, like)
	if err != nil {
		Error(w, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		var id, typ, src string
		var importedAt sql.NullTime
		if err := rows.Scan(&id, &typ, &src, &importedAt); err != nil {
			Error(w, http.StatusInternalServerError, err)
			return
		}
		results = append(results, map[string]any{
			"id":          id,
			"type":        typ,
			"source_file": src,
			"imported_at": importedAt.Time,
		})
	}
	JSON(w, http.StatusOK, map[string]any{
		"query":   url.QueryEscape(query),
		"results": results,
	})
}

//
// ============================================================
//  CORE ENTRY OPERATIONS (get/delete/reimport)
// ============================================================
//

// handleBootstrapGet — serves a bootstrap YAML file by ID.
func (s *Server) handleBootstrapGet(w http.ResponseWriter, id string, db *sql.DB) {
	// Try database first
	var content []byte
	if row := db.QueryRow(`SELECT content FROM bootstrap WHERE id = $1;`, id); row != nil {
		if err := row.Scan(&content); err == nil && len(content) > 0 {
			w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(content)
			return
		}
	}

	// If not found in DB, fallback to local file system
	basePath := "."
	if s.cfg != nil && s.cfg.RepoRoot != "" {
		basePath = s.cfg.RepoRoot
	}
	path := filepath.Join(basePath, "disyaml", "_bootstrap", id)
	if !strings.HasSuffix(path, ".yaml") {
		path += ".yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		Error(w, http.StatusNotFound, fmt.Errorf("bootstrap entry not found: %s", id))
		return
	}

	// ✅ Serve pure YAML — no JSON wrapping
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func deleteBootstrapEntry(w http.ResponseWriter, id string, db *sql.DB) {
	_, err := db.Exec(`DELETE FROM bootstrap WHERE id = $1;`, id)
	if err != nil {
		Error(w, http.StatusInternalServerError, err)
		return
	}
	JSON(w, http.StatusOK, map[string]string{"deleted": id})
}

func reimportBootstrap(w http.ResponseWriter, db *sql.DB) {
	go func() {
		if err := s_bootstrapImport(db); err != nil {
			fmt.Println("⚠️  reimport error:", err)
		} else {
			fmt.Println("✅ Bootstrap reimport completed")
		}
	}()
	JSON(w, http.StatusOK, map[string]string{"status": "reimport started"})
}

// s_bootstrapImport — temporary shim to run a fresh bootstrap import.
func s_bootstrapImport(db *sql.DB) error {
	cmd := exec.Command("go", "run", "./cmd/dis-core", "--bootstrap-only")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

//
// ============================================================
//  RECONCILE ENDPOINT (Finagler → DIS-Core bridge)
// ============================================================
//

type BootstrapReconcileRequest struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Version string `json:"version"`
	Domain  string `json:"domain"`
	Content string `json:"content"`
}

func (s *Server) handleBootstrapReconcile(w http.ResponseWriter, r *http.Request) {
	var req BootstrapReconcileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, fmt.Errorf("decode: %v", err))
		return
	}
	if req.ID == "" || req.Content == "" {
		Error(w, http.StatusBadRequest, fmt.Errorf("missing id or content"))
		return
	}

	basePath := "."
	if s.cfg != nil && s.cfg.RepoRoot != "" {
		basePath = s.cfg.RepoRoot
	}
	root := filepath.Join(basePath, "disyaml", "_bootstrap")
	if err := os.MkdirAll(root, 0755); err != nil {
		Error(w, http.StatusInternalServerError, fmt.Errorf("mkdir: %v", err))
		return
	}

	safeName := fmt.Sprintf("%s_%s_%s.yaml",
		req.Type, req.Domain, time.Now().Format("20060102_150405"))
	target := filepath.Join(root, safeName)

	if err := ioutil.WriteFile(target, []byte(req.Content), 0644); err != nil {
		Error(w, http.StatusInternalServerError, fmt.Errorf("write file: %v", err))
		return
	}

	resp := map[string]string{
		"status":   "ok",
		"message":  "bootstrap saved successfully",
		"filename": filepath.Base(target),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
