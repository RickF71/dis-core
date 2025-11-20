package service

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dis-core/cmd/dis-core/bootstrap"
	coreauth "dis-core/internal/core/authority"
	"dis-core/internal/policy"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StartDaemon starts the DIS-Core service as a daemon with graceful shutdown
func StartDaemon(cfg *bootstrap.Config, db *pgxpool.Pool, policyEngine *policy.OPAEngine, console *bootstrap.ConsoleComponents, engine *coreauth.Engine) error {
	log.Println("[daemon] Starting DIS-Core daemon...")

	// Create context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Initialize HTTP server components (reuse existing bootstrap logic)
	serverComponents, err := bootstrap.InitializeHTTPServer(db, nil, cfg, engine) // ledger will be nil for now
	if err != nil {
		return err
	}

	// Create HTTP server with timeouts
	server := &http.Server{
		Addr:         cfg.Address(),
		Handler:      serverComponents.Handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("[daemon] HTTP server listening on %s", cfg.Address())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[daemon] HTTP server error: %v", err)
		}
	}()

	log.Println("[daemon] DIS-Core daemon started successfully")

	// Wait for shutdown signal
	<-ctx.Done()

	log.Println("[daemon] Received shutdown signal, shutting down gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[daemon] HTTP server shutdown error: %v", err)
	} else {
		log.Println("[daemon] HTTP server stopped")
	}

	// Close database connections
	if db != nil {
		db.Close()
		log.Println("[daemon] Database connections closed")
	}

	log.Println("[daemon] DIS-Core daemon stopped gracefully")
	return nil
}

// StartDaemonLegacy keeps the old 4-arg signature used by existing tests and
// convenience callers. It constructs a minimal authority Engine and calls the
// canonical StartDaemon. This preserves backward compatibility so older tests
// and call sites that haven't been updated yet will compile and run.
func StartDaemonLegacy(cfg *bootstrap.Config, db *pgxpool.Pool, policyEngine *policy.OPAEngine, console *bootstrap.ConsoleComponents) error {
	// Create a minimal authority engine with the provided DB. Use the
	// engine's NewEngine helper from coreauth to avoid duplicating setup.
	eng := coreauth.NewEngine(&coreauth.Config{}, db)

	return StartDaemon(cfg, db, policyEngine, console, eng)
}

// Preserve a simple alias for older call sites/tests expecting a different
// exported symbol name.
var StartDaemonOld = StartDaemonLegacy
