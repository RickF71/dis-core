package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
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

	// --- Commit (finalize a domain into DB)
	mux.HandleFunc("/api/bootstrap/commit-domain/", withPOST(func(w http.ResponseWriter, r *http.Request) {
		s.handleBootstrapCommitDomain(w, r)
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
func (s *Server) handleBootstrapCommitDomain(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/bootstrap/commit-domain/")
	if id == "" {
		Error(w, http.StatusBadRequest, fmt.Errorf("missing id"))
		return
	}

	// Parse posted fields
	var req struct {
		Name   string `json:"name"`
		Parent string `json:"parent"`
		Schema string `json:"schema"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, fmt.Errorf("decode: %v", err))
		return
	}

	if req.Name == "" {
		Error(w, http.StatusBadRequest, fmt.Errorf("missing name"))
		return
	}

	// Provide fallbacks
	if req.Parent == "" {
		req.Parent = "domain.null"
	}
	if req.Schema == "" {
		req.Schema = "domain.void"
	}

	// Build fully qualified domain name (FDN)
	fdn := fmt.Sprintf("%s.%s", req.Parent, req.Name)

	// Ensure the domain name doesn’t already exist
	var exists bool
	if err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM domains WHERE name=$1)`, req.Name).Scan(&exists); err != nil {
		Error(w, http.StatusInternalServerError, fmt.Errorf("check exists: %v", err))
		return
	}
	if exists {
		Error(w, http.StatusConflict, fmt.Errorf("domain already exists: %s", req.Name))
		return
	}

	// Lookup parent_id (nullable)
	var parentID *string
	err := s.db.QueryRow(`SELECT id::text FROM domains WHERE name=$1`, req.Parent).Scan(&parentID)
	if err == sql.ErrNoRows {
		parentID = nil // top-level domain
	} else if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Errorf("find parent: %v", err))
		return
	}

	// Prepare minimal domain JSON for the 'data' column
	domainJSON := map[string]any{
		"meta": map[string]any{
			"name":      req.Name,
			"domain_id": fdn,
			"schema":    req.Schema,
		},
		"parents": []string{req.Parent},
	}
	dataBytes, _ := json.Marshal(domainJSON)

	// Insert new domain row
	_, err = s.db.Exec(`
		INSERT INTO domains (parent_id, name, data)
		VALUES (
			(SELECT id FROM domains WHERE name=$1),
			$2,
			$3::jsonb
		)`,
		req.Parent, req.Name, string(dataBytes))
	if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Errorf("insert domain: %v", err))
		return
	}

	// Remove from bootstrap after commit
	if _, err := s.db.Exec(`DELETE FROM bootstrap WHERE id=$1`, id); err != nil {
		Error(w, http.StatusInternalServerError, fmt.Errorf("delete bootstrap: %v", err))
		return
	}

	// Respond
	JSON(w, http.StatusOK, map[string]any{
		"status":    "reconciled",
		"domain_id": fdn,
		"name":      req.Name,
		"parent":    req.Parent,
		"schema":    req.Schema,
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
	path := filepath.Join(basePath, "_bootstrap", id)
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
	root := filepath.Join(basePath, "_bootstrap")
	if err := os.MkdirAll(root, 0755); err != nil {
		Error(w, http.StatusInternalServerError, fmt.Errorf("mkdir: %v", err))
		return
	}

	// --- Convert JSON content to YAML before writing ---
	var intermediate any
	if err := json.Unmarshal([]byte(req.Content), &intermediate); err != nil {
		Error(w, http.StatusBadRequest, fmt.Errorf("unmarshal json: %v", err))
		return
	}

	yamlBytes, err := yaml.Marshal(intermediate)
	if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Errorf("marshal yaml: %v", err))
		return
	}

	safeName := fmt.Sprintf("%s_%s_%s.yaml",
		req.Type, req.Domain, time.Now().Format("20060102_150405"))
	target := filepath.Join(root, safeName)

	if err := os.WriteFile(target, yamlBytes, 0644); err != nil {
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

type canonizeReq struct {
	Filename string `json:"filename"`
}

// POST /api/bootstrap/canonize
func (s *Server) handleBootstrapCanonize(w http.ResponseWriter, r *http.Request) {
	var req canonizeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Filename == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Flexible filename lookup
	var contentBytes []byte
	err := s.Ledger.DB.QueryRow(`
		SELECT content
		FROM bootstrap_files
		WHERE filename = $1
		   OR rel_path LIKE '%' || $1
		   OR $1 LIKE '%' || filename
		LIMIT 1
	`, req.Filename).Scan(&contentBytes)
	if err != nil {
		http.Error(w, "bootstrap file not found", http.StatusNotFound)
		return
	}

	// Convert YAML → JSON (with key normalization)
	var intermediate any
	if err := yaml.Unmarshal(contentBytes, &intermediate); err != nil {
		http.Error(w, "invalid YAML: "+err.Error(), http.StatusBadRequest)
		return
	}
	clean := yamlToJSON(intermediate) // 👈 normalize map keys
	jsonBytes, err := json.Marshal(clean)
	if err != nil {
		http.Error(w, "cannot convert to JSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// detect type (simple heuristic)
	ctype := "domain"
	if len(req.Filename) >= 7 && req.Filename[:7] == "schema." {
		ctype = "schema"
	}

	// Upsert into canon
	_, err = s.Ledger.DB.Exec(`
		INSERT INTO canon (type, content)
		VALUES ($1, $2::jsonb)
		ON CONFLICT DO NOTHING
	`, ctype, string(jsonBytes))
	if err != nil {
		http.Error(w, "canon insert failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Write canonical file to disk
	destDir := filepath.Join("disyaml", "canon")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		http.Error(w, "cannot create canon dir: "+err.Error(), http.StatusInternalServerError)
		return
	}
	dest := filepath.Join(destDir, req.Filename)
	if err := os.WriteFile(dest, contentBytes, 0o644); err != nil {
		http.Error(w, "cannot write canon file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"canonized","file":"` + req.Filename + `"}`))
}

// yamlToJSON recursively converts YAML maps with interface{} keys into JSON-safe maps
func yamlToJSON(v any) any {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		m2 := make(map[string]any)
		for k, v2 := range val {
			m2[castToString(k)] = yamlToJSON(v2)
		}
		return m2
	case map[string]interface{}:
		m2 := make(map[string]any)
		for k, v2 := range val {
			m2[k] = yamlToJSON(v2)
		}
		return m2
	case []interface{}:
		for i, v2 := range val {
			val[i] = yamlToJSON(v2)
		}
		return val
	default:
		return val
	}
}

func castToString(v any) string {
	switch s := v.(type) {
	case string:
		return s
	default:
		return fmt.Sprintf("%v", s)
	}
}
