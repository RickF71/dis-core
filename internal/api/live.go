package api

import (
	"encoding/json"
	"net/http"
	"time"

	"dis-core/internal/model"
)

// handleEventsLive streams mock events as Server-Sent Events.
func (s *Server) handleEventsLive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case t := <-ticker.C:
			ev := model.DisEvent{
				ID:        0,
				TS:        t,
				Type:      model.DisEventType("domain.freeze.v1"),
				Actor:     "domain.usa",
				Payload:   []byte(`{"demo":"tick"}`),
				Signature: "none",
			}

			data, _ := json.Marshal(ev)
			w.Write([]byte("data: "))
			w.Write(data)
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}
