package server

import (
	"database/sql"
	"net/http"

	"dis-core/internal/identity"
	"dis-core/internal/ledger"
	"dis-core/internal/schema"
)

// Server represents a minimal server structure for the server package
type Server struct {
	DB        *sql.DB
	Ledger    *ledger.Ledger
	Mux       *http.ServeMux
	Schemas   *schema.Registry
	TriadRepo *identity.TriadRepository
}

// New creates a new server instance
func New(db *sql.DB, led *ledger.Ledger) *Server {
	return &Server{
		DB:     db,
		Ledger: led,
		Mux:    http.NewServeMux(),
	}
}

// Handler returns the HTTP mux
func (s *Server) Handler() *http.ServeMux {
	return s.Mux
}
