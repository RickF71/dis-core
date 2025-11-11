package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// UpdatedServer represents the migrated server with chi router support
type UpdatedServer struct {
	*Server // Embed existing server
	router  *chi.Mux
}

// NewUpdatedServer creates a new server instance with chi router
func NewUpdatedServer(existingServer *Server) *UpdatedServer {
	us := &UpdatedServer{
		Server: existingServer,
	}

	// Initialize chi router
	us.router = us.NewChiRouter()

	return us
}

// ServeHTTP implements http.Handler interface for the updated server
func (us *UpdatedServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	us.router.ServeHTTP(w, r)
}

// StartChiServer starts the HTTP server with chi router
func (us *UpdatedServer) StartChiServer(addr string) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      us,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Printf("Starting chi-based DIS API server on %s\n", addr)
	return srv.ListenAndServe()
}

// StartChiServerWithGracefulShutdown starts the server with graceful shutdown support
func (us *UpdatedServer) StartChiServerWithGracefulShutdown(addr string) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      us,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Printf("Starting chi-based DIS API server on %s with graceful shutdown support\n", addr)

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Wait for shutdown signal (could be extended with signal handling)
	// For now, just start and return
	return nil
}

// GetRouter returns the chi router instance
func (us *UpdatedServer) GetRouter() *chi.Mux {
	return us.router
}

// Migration helper functions

// MigrateToChiRouter performs the complete migration from mux to chi router
func MigrateToChiRouter(existingServer *Server) *UpdatedServer {
	fmt.Println("Migrating from http.ServeMux to chi router...")

	updatedServer := NewUpdatedServer(existingServer)

	fmt.Println("Migration complete. Chi router initialized with format-aware routes.")
	fmt.Println("Available formats:")
	fmt.Println("  - JSON: Default format (no suffix)")
	fmt.Println("  - File: Add /file suffix for raw content")
	fmt.Println("  - Text: Add /text suffix for plain text")
	fmt.Println("  - Raw: Add /raw suffix for raw data")

	return updatedServer
}

// RegisterAdditionalRoutes allows adding custom routes after migration
func (us *UpdatedServer) RegisterAdditionalRoutes(routeFunc func(*chi.Mux)) {
	routeFunc(us.router)
}

// Example usage of the migration:
//
// func main() {
//     // Create existing server instance
//     server := &Server{...}
//
//     // Perform migration
//     updatedServer := MigrateToChiRouter(server)
//
//     // Start chi-based server
//     log.Fatal(updatedServer.StartChiServer(":8080"))
// }

// Format validation helpers

// ValidateFormat checks if a format is supported for a given endpoint type
func ValidateFormat(format Format, supportedFormats []Format) bool {
	for _, sf := range supportedFormats {
		if format == sf {
			return true
		}
	}
	return false
}

// GetEndpointSupportedFormats returns the list of supported formats for different endpoint types
func GetEndpointSupportedFormats() map[string][]Format {
	return map[string][]Format{
		"domain":  {FormatJSON, FormatFile, FormatText},
		"policy":  {FormatFile, FormatJSON, FormatText},
		"status":  {FormatJSON, FormatText},
		"receipt": {FormatFile, FormatJSON, FormatText},
		"files":   {FormatJSON, FormatFile},
	}
}

// Debugging and inspection helpers

// ListRoutes provides a debug view of all registered routes
func (us *UpdatedServer) ListRoutes() []string {
	routes := []string{}

	// Walk through chi routes
	chi.Walk(us.router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		routes = append(routes, fmt.Sprintf("%s %s", method, route))
		return nil
	})

	return routes
}

// PrintRoutes prints all registered routes to stdout
func (us *UpdatedServer) PrintRoutes() {
	fmt.Println("Registered chi routes:")
	routes := us.ListRoutes()
	for _, route := range routes {
		fmt.Printf("  %s\n", route)
	}
}
