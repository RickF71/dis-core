package api

import (
	"encoding/json"
	"net/http"

	"github.com/lib/pq"

	"dis-core/internal/jikka" // where your Jikka struct and logic live
)

// GET /api/jikka?id=<uuid>
func (s *Server) handleGetJikka(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}
	// FIXME: jikka.GetJikkaByID expects *sql.DB but we have *pgx.Conn
	// Need to migrate internal/jikka package to pgx
	// id, err := uuid.Parse(idStr)
	// if err != nil {
	// 	http.Error(w, "invalid id", http.StatusBadRequest)
	// 	return
	// }
	// j, err := jikka.GetJikkaByID(s.DB(), id)
	// if err != nil {
	// 	http.Error(w, "jikka not found", http.StatusNotFound)
	// 	return
	// }
	//
	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(j)

	http.Error(w, "Jikka handlers temporarily disabled during pgx migration", http.StatusNotImplemented)
}

// POST /api/jikka
func (s *Server) handleCreateJikka(w http.ResponseWriter, r *http.Request) {
	var j jikka.Jikka
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if len(j.Participants) < 2 {
		http.Error(w, "need at least two participants", http.StatusBadRequest)
		return
	}

	// FIXME: jikka.CreateJikka expects *sql.DB but we have *pgx.Conn
	// Need to migrate internal/jikka package to pgx
	// created, err := jikka.CreateJikka(s.DB(), &j)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	//
	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(created)

	http.Error(w, "Jikka creation temporarily disabled during pgx migration", http.StatusNotImplemented)
}

// GET /api/jikka/list
func (s *Server) handleListJikkas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Ensure DB is available for this handler; listing requires a DB.
	db := s.requireDB(w)
	if db == nil {
		return
	}

	rows, err := db.Query(ctx, `SELECT id, participants, type, flows, rules, state, created_at, updated_at FROM jikkas ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []jikka.Jikka
	for rows.Next() {
		var j jikka.Jikka
		if err := rows.Scan(&j.ID, pq.Array(&j.Participants), &j.Flows, &j.Rules, &j.State, &j.CreatedAt, &j.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, j)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}
