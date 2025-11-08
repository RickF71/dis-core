package main

import (
	"context"
	"log"
	"net/http"
	"time"

	apiserver "dis-core/internal/api/server"
	"dis-core/internal/db"
)

func buildServer() *http.ServeMux {
	store := db.DefaultConn
	if store == nil {
		log.Fatal("database not initialized")
	}

	// Create server with database (ledger will be nil for now)
	s := apiserver.New(store, nil)

	// The NewServer call already registers all routes and sets up the mux
	return s.Handler() // Updated method name
}

// RunServer starts the HTTP server and listens until context cancel.
func RunServer(ctx context.Context, addr string) {
	server := &http.Server{
		Addr:         addr,
		Handler:      buildServer(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🌍 DIS-CORE API listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("🛑 shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
