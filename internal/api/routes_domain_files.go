package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// --------------------------------------------------------
// Domain Files API (JSONB-backed, atomic writes per file)
// --------------------------------------------------------

// GET /api/domain/{id}/files  →  list file names
func (s *Server) handleDomainFilesList(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := r.Context()

	q := `SELECT jsonb_object_keys(files) FROM domains WHERE id = $1`
	rows, err := s.DB().Query(ctx, q, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			files = append(files, name)
		}
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]any{"files": files}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GET /api/domain/{id}/file/{filename}  →  return file content as text
func (s *Server) handleDomainFileGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	ctx := r.Context()

	//var content sql.NullString
	q := `SELECT files->>$2 FROM domains WHERE id=$1`
	// ↑ this pulls the full JSON string for that key (the whole object)
	var raw sql.NullString
	err := s.DB().QueryRow(ctx, q, id, filename).Scan(&raw)
	if err == sql.ErrNoRows {
		http.Error(w, "domain not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !raw.Valid {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	// Unmarshal the JSON object to get just "content"
	var obj map[string]any
	if err := json.Unmarshal([]byte(raw.String), &obj); err != nil {
		http.Error(w, "invalid json stored", http.StatusInternalServerError)
		return
	}
	val, _ := obj["content"].(string)

	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, val)
}

// PUT /api/domain/{id}/file/{filename}  →  create or overwrite
func (s *Server) handleDomainFilePut(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	ctx := r.Context()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	content := string(data)

	q := `
	  UPDATE domains
	  SET files = jsonb_set(
	    COALESCE(files, '{}'::jsonb),
	    ARRAY[$2]::text[],
	    jsonb_build_object(
	      'author', COALESCE(files->$2->>'author', 'system'),
	      'content', $3::text,          -- ✅ explicit type cast fixes it
	      'updated_at', now()
	    ),
	    true
	  )
	  WHERE id = $1::uuid
	`
	if _, err := s.DB().Exec(ctx, q, id, filename, content); err != nil {
		http.Error(w, "db update failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

// DELETE /api/domain/{id}/file/{filename}
func (s *Server) handleDomainFileDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	ctx := r.Context()

	q := `UPDATE domains SET files = files - $2 WHERE id=$1::uuid`
	if _, err := s.DB().Exec(ctx, q, id, filename); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// POST /api/domain/{id}/file/rename  →  rename an existing file
func (s *Server) handleDomainFileRename(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := r.Context()

	var req struct {
		Old string `json:"old"`
		New string `json:"new"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Old == "" || req.New == "" {
		http.Error(w, "missing filenames", http.StatusBadRequest)
		return
	}

	q := `
	UPDATE domains
	SET files = jsonb_set(
		files - $2,
		ARRAY[$3],
		files->$2
	)
	WHERE id = $1::uuid
	`
	if _, err := s.DB().Exec(ctx, q, id, req.Old, req.New); err != nil {
		http.Error(w, "rename failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "renamed",
		"oldName": req.Old,
		"newName": req.New,
	})
}

// POST /api/domain/{id}/file/{filename}  →  create new only if not exists
func (s *Server) handleDomainFileCreate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	content := string(body)

	q := `
	UPDATE domains
	SET files = jsonb_set(
		COALESCE(files, '{}'::jsonb),
		ARRAY[$2],
		CASE
			WHEN files ? $2 THEN files->$2
			ELSE jsonb_build_object(
				'author', 'system',
				'content', $3,
				'updated_at', now()
			)
		END,
		true
	)
	WHERE id=$1::uuid
	RETURNING (files->$2 IS NOT NULL) AS existed
	`

	var existed sql.NullBool
	if err := s.DB().QueryRow(ctx, q, id, filename, content).Scan(&existed); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existed.Valid && existed.Bool {
		http.Error(w, "file already exists", http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
