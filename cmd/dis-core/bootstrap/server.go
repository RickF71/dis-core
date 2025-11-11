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

	// Initialize policy engine
	policyComponents, err := InitializePolicyEngine()
	if err != nil {
		log.Printf("⚠️  Warning: Could not initialize policy engine: %v", err)
		// Continue without policy engine - use nil
		policyComponents = nil
	}

	// Create API server with policy engine
	var srv *api.Server
	if policyComponents != nil {
		srv = api.NewWithPolicy(database, ledger, policyComponents.Engine)
		log.Println("✅ API server initialized with policy engine")
	} else {
		srv = api.New(database, ledger)
		log.Println("✅ API server initialized without policy engine")
	}

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
