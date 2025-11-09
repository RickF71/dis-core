// server.go — root package main
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"dis-core/internal/api"
	"dis-core/internal/db"
	"dis-core/internal/ledger"
)

// buildServer initializes and returns the main HTTP mux for DIS-Core.
func buildServer() *http.ServeMux {
	// Initialize the connection pool.
	store := db.DefaultConn
	if store == nil {
		log.Fatal("database not initialized (db.DefaultConn is nil)")
	}

	// Initialize the ledger (may evolve to use store + context)
	led := &ledger.Ledger{DB: store}

	// Create the main API server instance.
	s := api.New(store, led)

	// The API server wires routes internally and exposes the mux.
	return s.Handler()
}

// RunServer starts the DIS-Core HTTP server and gracefully shuts it down on context cancel.
func RunServer(ctx context.Context, addr string) {
	server := &http.Server{
		Addr:         addr,
		Handler:      buildServer(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🌍 DIS-Core API listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("🛑 Shutting down DIS-Core server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
