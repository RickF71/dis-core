package api

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/events"
)

func (s *Server) PostEvent(w http.ResponseWriter, r *http.Request) {
	var e events.DisEvent
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	_, err := s.DB().Exec(
		`INSERT INTO events (type, actor, payload, sig) VALUES ($1,$2,$3,$4)`,
		e.Type, e.Actor, e.Payload, e.Signature,
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(201)
}
