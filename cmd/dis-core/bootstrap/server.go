package bootstrap

import (
	"log"
	"net/http"

	"dis-core/internal/api"
	"dis-core/internal/api/server"
	"dis-core/internal/ledger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ServerComponents holds the initialized HTTP server components
type ServerComponents struct {
	Server  *api.Server
	Handler http.Handler
	Address string
}

// InitializeHTTPServer creates and configures the HTTP API server
func InitializeHTTPServer(database *pgxpool.Pool, ledger *ledger.Ledger, config *Config) (*ServerComponents, error) {
	log.Printf("🚀 DIS-Core %s starting on http://%s", config.Version, config.Address())

	// Create API server
	srv := api.New(database, ledger)

	// Wrap with CORS middleware
	handler := server.WithCORS(srv.Handler())

	return &ServerComponents{
		Server:  srv,
		Handler: handler,
		Address: config.Address(),
	}, nil
}

// StartServer starts the HTTP server and blocks until it exits
func (sc *ServerComponents) StartServer() error {
	log.Printf("🌐 Server listening on %s", sc.Address)
	return http.ListenAndServe(sc.Address, sc.Handler)
}
