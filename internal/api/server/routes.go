package server

import (
	authorityapi "dis-core/internal/api/authority"
)

// RegisterRoutes wires all API endpoints.
func (s *Server) RegisterRoutes() {
	s.Mux.HandleFunc("/api/health", s.HandleHealth)
	s.Mux.HandleFunc("/api/events/live", s.handleEventsLive)

	// GOV-2: Authority Console endpoints (read-only)
	if s.TriadRepo != nil {
		triadHandler := authorityapi.NewTriadHandler(s.TriadRepo)
		s.Mux.HandleFunc("/api/authority/triad/", triadHandler.GetTriadByIdentity)
		s.Mux.HandleFunc("/api/authority/flow/preview", triadHandler.PreviewFlow)
		s.Mux.HandleFunc("/api/authority/flow/eval", triadHandler.EvaluateFlow)
	}
}
