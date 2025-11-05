package events

import (
	"dis-core/internal/model"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HandleEventsLive streams domain events as SSE for Finagler’s live map
func HandleEventsLive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Demo: alternate between freeze / unfreeze every few seconds
	states := []model.DisEventType{"domain.freeze.v1", "domain.unfreeze.v1"}
	idx := 0

	for {
		ev := model.DisEvent{
			ID:        idx,
			TS:        time.Now(),
			Type:      states[idx%len(states)],
			Actor:     "domain.usa",
			Payload:   []byte(fmt.Sprintf(`{"domain":"domain.usa","state":"%s"}`, states[idx%len(states)])),
			Signature: "",
		}

		bytes, _ := json.Marshal(ev)
		fmt.Fprintf(w, "event: domain.update\n")
		fmt.Fprintf(w, "data: %s\n\n", string(bytes))
		flusher.Flush()

		idx++
		time.Sleep(4 * time.Second)
	}
}
