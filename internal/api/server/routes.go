package server

// RegisterRoutes wires all API endpoints.
func (s *Server) RegisterRoutes() {
	s.Mux.HandleFunc("/api/health", s.HandleHealth)
	s.Mux.HandleFunc("/api/events/live", s.handleEventsLive)
}
