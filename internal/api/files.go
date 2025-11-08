package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) registerFileRoutes() {
	mux := s.mux

	// Specific routes first
	mux.HandleFunc("/api/files/search", s.handleFileSearch)

	mux.HandleFunc("/api/files/export", s.handleFileExport)

	// Then by-ID (prefix)
	mux.HandleFunc("/api/files/", s.handleFileByID)

	// Then the base list
	mux.HandleFunc("/api/files", s.handleListFiles)

	fmt.Println("✅ Virtual file API registered")
}

// ------------------------------------------------------------
//  HANDLERS
// ------------------------------------------------------------

// handleListFiles — GET /api/files
// Returns list of all imported files (id, rel_path, filename, size, exported_to)
func (s *Server) handleListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := s.DB().Query(`
		SELECT id, rel_path, filename,
		       octet_length(content) AS size,
		       COALESCE(exported_to, '') AS exported_to,
		       imported_at, exported_at
		FROM bootstrap_files
		ORDER BY rel_path ASC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []map[string]any
	for rows.Next() {
		var id int
		var relPath, filename, exportedTo string
		var size int
		var importedAt, exportedAt sql.NullTime

		if err := rows.Scan(&id, &relPath, &filename, &size, &exportedTo, &importedAt, &exportedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		files = append(files, map[string]any{
			"id":          id,
			"rel_path":    relPath,
			"filename":    filename,
			"size":        size,
			"exported_to": exportedTo,
			"imported_at": importedAt.Time,
			"exported_at": exportedAt.Time,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"count": len(files),
		"items": files,
	})
}

// handleFileByID — GET or PUT /api/files/:id
// GET  → returns file contents
// PUT  → updates file content (Base64 or raw text)
func (s *Server) handleFileByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/files/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid file id", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		var filename, relPath string
		var content []byte

		if err := s.DB().QueryRow(`
			SELECT filename, rel_path, content
			FROM bootstrap_files WHERE id = $1
		`, id).Scan(&filename, &relPath, &content); err != nil {
			if err == sql.ErrNoRows {
				http.NotFound(w, r)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Serve raw content (binary safe)
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s", filename))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)

	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Allow either base64 JSON or plain text
		var decoded []byte
		if json.Valid(body) {
			var obj map[string]string
			if err := json.Unmarshal(body, &obj); err == nil && obj["content"] != "" {
				decoded, err = base64.StdEncoding.DecodeString(obj["content"])
				if err != nil {
					http.Error(w, "base64 decode: "+err.Error(), http.StatusBadRequest)
					return
				}
			}
		}
		if len(decoded) == 0 {
			decoded = body
		}

		_, err = s.DB().Exec(`UPDATE bootstrap_files SET content = $1 WHERE id = $2`, decoded, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "file updated successfully",
		})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleFileSearch — GET /api/files/search?q=term
// Performs a case-insensitive LIKE search on rel_path and filename.
func (s *Server) handleFileSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		http.Error(w, "missing search query ?q=", http.StatusBadRequest)
		return
	}

	like := "%" + strings.ToLower(query) + "%"
	rows, err := s.DB().Query(`
		SELECT id, rel_path, filename,
		       octet_length(content) AS size,
		       COALESCE(exported_to, '') AS exported_to,
		       imported_at, exported_at
		FROM bootstrap_files
		WHERE LOWER(rel_path) LIKE $1 OR LOWER(filename) LIKE $1
		ORDER BY rel_path ASC
	`, like)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []map[string]any
	for rows.Next() {
		var id int
		var relPath, filename, exportedTo string
		var size int
		var importedAt, exportedAt sql.NullTime
		if err := rows.Scan(&id, &relPath, &filename, &size, &exportedTo, &importedAt, &exportedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		files = append(files, map[string]any{
			"id":          id,
			"rel_path":    relPath,
			"filename":    filename,
			"size":        size,
			"exported_to": exportedTo,
			"imported_at": importedAt.Time,
			"exported_at": exportedAt.Time,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"count": len(files),
		"items": files,
	})
}

// handleFileExport — POST /api/files/export
// Marks one or more files as exported (updates exported_to and exported_at).
func (s *Server) handleFileExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		IDs        []int  `json:"ids"`         // optional: list of IDs to mark
		RelPath    string `json:"rel_path"`    // or a single path
		ExportedTo string `json:"exported_to"` // where it was exported (canon path, freeze, etc.)
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if len(req.IDs) == 0 && req.RelPath == "" {
		http.Error(w, "missing ids[] or rel_path", http.StatusBadRequest)
		return
	}
	if req.ExportedTo == "" {
		req.ExportedTo = "canon"
	}

	var result sql.Result
	var err error

	if len(req.IDs) > 0 {
		// Batch update by IDs
		placeholders := make([]string, len(req.IDs))
		args := make([]interface{}, len(req.IDs)+1)
		for i, id := range req.IDs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = id
		}
		args[len(args)-1] = req.ExportedTo

		q := fmt.Sprintf(`
			UPDATE bootstrap_files
			SET exported_to = $%d, exported_at = NOW()
			WHERE id IN (%s) AND exported_to IS NULL
		`, len(args), strings.Join(placeholders, ","))
		result, err = s.DB().Exec(q, args...)
	} else {
		// Single update by rel_path
		result, err = s.DB().Exec(`
			UPDATE bootstrap_files
			SET exported_to = $1, exported_at = NOW()
			WHERE rel_path = $2 AND exported_to IS NULL
		`, req.ExportedTo, req.RelPath)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("update failed: %v", err), http.StatusInternalServerError)
		return
	}

	affected, _ := result.RowsAffected()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":      "ok",
		"exported_to": req.ExportedTo,
		"updated":     affected,
	})
}
