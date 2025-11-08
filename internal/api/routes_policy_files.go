package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
)

// PolicyRecord represents one row from the policies table.
type PolicyRecord struct {
	ID        uuid.UUID  `json:"id"`
	DomainID  *uuid.UUID `json:"domain_id,omitempty"`
	Name      string     `json:"name"`
	Content   string     `json:"content"`
	CreatedAt string     `json:"created_at"`
}

// RegisterPolicyFileRoutes adds CRUD endpoints for managing policy files.
func (s *Server) registerPolicyFileRoutes() {
	mux := s.mux
	mux.HandleFunc("GET /api/policy/list", s.handleListPolicies)
	mux.HandleFunc("GET /api/policy/{name}", s.handleGetPolicy)
	mux.HandleFunc("PUT /api/policy/{name}", s.handlePutPolicy)
	mux.HandleFunc("DELETE /api/policy/{name}", s.handleDeletePolicy)
}

// GET /api/policy/list
func (s *Server) handleListPolicies(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(`SELECT id, domain_id, name, content, created_at FROM policies ORDER BY name`)
	if err != nil {
		http.Error(w, "failed to list policies: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var out []PolicyRecord
	for rows.Next() {
		var p PolicyRecord
		if err := rows.Scan(&p.ID, &p.DomainID, &p.Name, &p.Content, &p.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out = append(out, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// GET /api/policy/{name}
func (s *Server) handleGetPolicy(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	row := s.db.QueryRow(`SELECT id, domain_id, name, content, created_at FROM policies WHERE name = $1`, name)
	var p PolicyRecord
	if err := row.Scan(&p.ID, &p.DomainID, &p.Name, &p.Content, &p.CreatedAt); err != nil {
		http.Error(w, "policy not found: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// PUT /api/policy/{name}
func (s *Server) handlePutPolicy(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	body, _ := io.ReadAll(r.Body)
	content := string(body)

	_, err := s.db.Exec(`
		INSERT INTO policies (name, content)
		VALUES ($1, $2)
		ON CONFLICT (name) DO UPDATE SET content = EXCLUDED.content, created_at = NOW()
	`, name, content)
	if err != nil {
		http.Error(w, "failed to save policy: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// DELETE /api/policy/{name}
func (s *Server) handleDeletePolicy(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	_, err := s.db.Exec(`DELETE FROM policies WHERE name = $1`, name)
	if err != nil {
		http.Error(w, "failed to delete policy: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"deleted"}`))
}
