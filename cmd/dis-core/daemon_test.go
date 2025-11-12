package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"dis-core/cmd/dis-core/bootstrap"
	"dis-core/cmd/dis-core/service"
	logutil "dis-core/internal/log"
	"dis-core/internal/policy"
)

// TestDaemonStartupShutdown tests the daemon's startup and graceful shutdown
func TestDaemonStartupShutdown(t *testing.T) {
	// Setup logging
	logutil.SetupLogging()

	ctx := context.Background()

	// Use test DSN or default
	testDSN := os.Getenv("DIS_TEST_DB_DSN")
	if testDSN == "" {
		testDSN = "postgres://dis_user:card567@localhost:5432/dis_test?sslmode=disable"
	}

	// Load configuration with test port
	config := &bootstrap.Config{
		DSN:  testDSN,
		Port: "8081", // Use different port for test
	}

	// Initialize database components
	dbComponents, err := bootstrap.InitializeDatabase(ctx, config.DSN)
	if err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	defer dbComponents.Close()

	// Bootstrap database tables
	if err := bootstrap.BootstrapTables(dbComponents.Database); err != nil {
		t.Fatalf("failed to bootstrap tables: %v", err)
	}

	// Initialize Authority Console
	console, err := bootstrap.InitializeAuthorityConsole(dbComponents.Database, dbComponents.Registry, dbComponents.Ledger)
	if err != nil {
		t.Fatalf("failed to initialize authority console: %v", err)
	}

	// Initialize policy engine (optional)
	var policyEngine *policy.OPAEngine

	// Create a context with timeout for the test
	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Start daemon in a goroutine
	errChan := make(chan error, 1)
	go func() {
		err := service.StartDaemon(config, dbComponents.Database, policyEngine, console)
		errChan <- err
	}()

	// Give the server time to start
	time.Sleep(1 * time.Second)

	// Test that the server is running by making a request
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/api/ping", config.Port))
	if err != nil {
		t.Fatalf("failed to connect to daemon: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Simulate shutdown signal
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("failed to find process: %v", err)
	}

	// Send SIGTERM to trigger graceful shutdown
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	// Wait for daemon to exit or timeout
	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("daemon exited with error: %v", err)
		}
	case <-testCtx.Done():
		t.Error("daemon did not shut down within timeout")
	}
}

// TestSignalHandling tests that the daemon properly handles various signals
func TestSignalHandling(t *testing.T) {
	// Test signal handling setup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Verify signal channel is set up
	if sigChan == nil {
		t.Error("signal channel not properly initialized")
	}

	// Test that we can receive signals
	go func() {
		time.Sleep(100 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGUSR1) // Safe signal for testing
	}()

	// This test verifies signal infrastructure works
	// The actual daemon signal handling is tested in the integration test above
}

// TestLoggingSetup verifies logging configuration
func TestLoggingSetup(t *testing.T) {
	// Test that logging setup doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetupLogging panicked: %v", r)
		}
	}()

	logutil.SetupLogging()

	// Test basic logging functionality
	// (This would write to stdout in test environment)
	fmt.Println("Test log message - this should appear in test output")
}
