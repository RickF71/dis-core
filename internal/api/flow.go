package api

import (
	"encoding/json"
	"net/http"
)

// ================================================
//  FLOW API — Alphatote / Chaostote rule endpoints
// ================================================

// registerFlowAPI adds all /api/flow/* routes to the server mux.
func registerFlowAPI(s *Server) {
	m := s.Mux()
	m.HandleFunc("/api/flow/rule", s.handleGetFlowRule)
	m.HandleFunc("/api/flow/rule/update", s.handleUpdateFlowRule)
	m.HandleFunc("/api/flow/status", s.handleFlowStatus)
	m.HandleFunc("/api/flow/simulate", s.handleFlowSimulate)
}

// Influence defines how a domain shapes Alphatote passing through it.
type Influence struct {
	Domain string  `json:"domain"`
	Kappa  float64 `json:"kappa"`
	Mode   string  `json:"mode"`
	Filter string  `json:"filter"`
}

// AlphaFlow describes an Alphatote flow path.
type AlphaFlow struct {
	Origin      string      `json:"origin"`
	Path        []Influence `json:"path"`
	Aggregation string      `json:"aggregation"`
	Threshold   float64     `json:"threshold"`
}

// handleGetFlowRule returns the canonical alphatote rule as JSON.
func (s *Server) handleGetFlowRule(w http.ResponseWriter, r *http.Request) {
	rule := AlphaFlow{
		Origin: "person",
		Path: []Influence{
			{"domain.personal", 0.94, "emotional", "balance_affection"},
			{"domain.government.usa", 0.76, "civic", "regulatory_damping"},
			{"domain.terra", 0.88, "ecological", "sustainability_bias"},
			{"domain.null", 1.00, "equilibrium", "identity_dissolution"},
		},
		Aggregation: "product",
		Threshold:   0.05,
	}
	writeJSON(w, http.StatusOK, rule)
}

// handleUpdateFlowRule accepts a new rule from Finagler (JSON body).
func (s *Server) handleUpdateFlowRule(w http.ResponseWriter, r *http.Request) {
	var f AlphaFlow
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// TODO: Persist rule to database or var/rules cache
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"saved":  true,
	})
}

// handleFlowStatus reports a summary of active flow coefficients.
func (s *Server) handleFlowStatus(w http.ResponseWriter, r *http.Request) {
	status := []map[string]any{
		{"domain": "personal", "kappa": 0.94, "mode": "emotional"},
		{"domain": "government", "kappa": 0.76, "mode": "civic"},
		{"domain": "terra", "kappa": 0.88, "mode": "ecological"},
	}
	writeJSON(w, http.StatusOK, status)
}

// handleFlowSimulate performs a simple propagation calculation.
// Finagler can POST {"base":1.0, "path":["personal","government","terra"]}
func (s *Server) handleFlowSimulate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Base float64  `json:"base"`
		Path []string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Temporary hardcoded kappas — replace with DB lookup
	kappas := map[string]float64{
		"personal":   0.94,
		"government": 0.76,
		"terra":      0.88,
	}

	mag := req.Base
	for _, d := range req.Path {
		if k, ok := kappas[d]; ok {
			mag *= k
		}
	}
	if mag < 0.05 {
		mag = 0
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"result":      mag,
		"base":        req.Base,
		"path":        req.Path,
		"aggregation": "product",
	})
}
