package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"dis-core/internal/api"
	"dis-core/internal/db"
)

// buildServer assembles all active routes.
func buildServer() *http.ServeMux {
	mux := http.NewServeMux()

	// === DIS-CORE v0.9.3 Routes ===
	api.RegisterConsoleAuthRoutes(mux)
	api.RegisterStatusRoutes(mux)

	// === Root route ===
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "🌐 DIS-CORE v0.9.3 — Self-Maintenance and Reflexive Identity\nTime: %s\n", db.NowRFC3339Nano())
	})

	return mux
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
