package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"dis-core/internal/api"
	"dis-core/internal/db"
)

type Server struct {
	store *sql.DB
	mux   *http.ServeMux
	// maybe cfg *config.Config
}

func buildServer() *http.ServeMux {
	store := db.DefaultDB
	if store == nil {
		log.Fatal("database not initialized")
	}

	// If you don't yet load a real config or policy, just pass nil and empty strings
	s := api.NewServer(store)

	// The NewServer call already registers all routes and sets up the mux
	return s.Mux() // We'll add this accessor next if it's missing
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
