package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// handlePing responds with a simple JSON heartbeat.
func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{
		"status":  "ok",
		"message": "DIS-Core alive",
		"time":    time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleOrphanInfo — return canonical orphan domain + schema info.
func (s *Server) handleOrphanInfo(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"domain_id":   "domain.null.orphans",
		"schema_id":   "domain.orphan.v1",
		"room_id":     "default",
		"description": "Default neutral space for unattached identities.",
		"rights": []string{
			"read.global.public",
			"post.orphan.self",
			"request.join.domain",
			"receive.civic.invite",
		},
		"obligations": []string{
			"must.acknowledge.code.of.conduct",
			"must.establish.binding.within.30d",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// assignDefaultDomainIfNone is a helper to ensure new identities
// are safely assigned to the orphan domain if unlinked.
type Person struct {
	ID        string `json:"id"`
	DomainID  string `json:"domain_id"`
	SchemaID  string `json:"schema_id"`
	RoomID    string `json:"room_id"`
	CreatedAt time.Time
}

const DefaultOrphanDomain = "domain.null.orphans"
const DefaultOrphanSchema = "domain.orphan.v1"
const DefaultOrphanRoom = "default"

func assignDefaultDomainIfNone(p *Person) {
	if p.DomainID == "" {
		p.DomainID = DefaultOrphanDomain
		p.SchemaID = DefaultOrphanSchema
		p.RoomID = DefaultOrphanRoom
	}
	fmt.Printf("[bootstrap] Assigned %s to orphan domain\n", p.ID)
}
