package api

import (
	"net/http"
	"strings"

	"dis-core/internal/domain"
)

// handleDomainInfo responds to /api/domain/info?code=XYZ
func (s *Server) handleDomainInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing 'code' query parameter"})
		return
	}

	// repo root: assume repo root (server runs in repo); change if you keep config field
	repoRoot := "."
	d, err := domain.LookupByCode(code, repoRoot)
	if err != nil {
		// inspect error prefix to determine status
		msg := err.Error()
		if strings.HasPrefix(msg, "400 ") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": strings.TrimPrefix(msg, "400 ")})
			return
		}
		if strings.HasPrefix(msg, "404 ") {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": strings.TrimPrefix(msg, "404 ")})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": msg})
		return
	}

	// Return the minimal domain fields as JSON
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"code":        d.Code,
		"name":        d.Name,
		"seat":        d.Seat,
		"lineage":     d.Lineage,
		"population":  d.Population,
		"description": d.Description,
	})
}
