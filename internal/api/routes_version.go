package api

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"
)

// These can be replaced at build time via ldflags, e.g.:
// go build -ldflags "-X dis-core/internal/api.buildVersion=v0.9.8 -X dis-core/internal/api.buildHash=$(git rev-parse --short HEAD)"
var (
	buildVersion  = "v0.9.8"
	buildHash     = "local"
	buildName     = "nightfield"
	bridgeVersion = "dis-bridge@0.9.3"
)

// VersionInfo provides structured metadata about the running DIS-Core instance.
type VersionInfo struct {
	Version  string `json:"version"`
	CoreHash string `json:"coreHash"`
	Build    string `json:"build"`
	Bridge   string `json:"bridge"`
	Schemas  int    `json:"schemas"`
	Domains  int    `json:"domains"`
	Time     string `json:"time"`
	Status   string `json:"status"`
}

// registerVersionRoutes exposes /api/version for Finagler and system introspection.
func (s *Server) registerVersionRoutes() {
	mux := s.mux

	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Pull commit info if buildHash == "local"
		commit := buildHash
		if commit == "local" {
			if info, ok := debug.ReadBuildInfo(); ok {
				commit = info.Main.Version
			}
		}

		info := VersionInfo{
			Version:  buildVersion,
			CoreHash: commit,
			Build:    buildName,
			Bridge:   bridgeVersion,
			Schemas:  s.schemaRegistryCount(),
			Domains:  s.domainRegistryCount(),
			Time:     time.Now().UTC().Format(time.RFC3339),
			Status:   "ok",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})
}

// Temporary helpers until SchemaRegistry and DomainRegistry expose counts.
func (s *Server) schemaRegistryCount() int { return 3 }
func (s *Server) domainRegistryCount() int { return 3 }
