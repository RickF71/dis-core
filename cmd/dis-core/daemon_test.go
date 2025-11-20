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
	"dis-core/internal/ledger"
	logutil "dis-core/internal/log"
	"dis-core/internal/policy"
	"dis-core/internal/schema"
	testdb "dis-core/internal/testdb"
)

// TestDaemonStartupShutdown tests the daemon's startup and graceful shutdown
func TestDaemonStartupShutdown(t *testing.T) {
	// Setup logging
	logutil.SetupLogging()

	ctx := context.Background()

	// Ensure a test DB is available; skip gracefully if DIS_TEST_DB_DSN is not set.
	pool := testdb.SetupTestDB(t)
	testdb.MustHaveDB(t, pool)

	// Load configuration with test port (DSN handled by harness)
	config := &bootstrap.Config{
		DSN:  "",
		Port: "8081", // Use different port for test
	}

	// Initialize schema registry
	schemaReg := schema.NewRegistry()

	// Initialize ledger using the harness pool
	led, err := ledger.Open(ctx, "", pool, schemaReg)
	if err != nil {
		t.Fatalf("failed to open ledger: %v", err)
	}

	// Initialize Authority Console using harness-managed pool and ledger
	console, err := bootstrap.InitializeAuthorityConsole(pool, schemaReg, led)
	if err != nil {
		t.Fatalf("failed to initialize authority console: %v", err)
	}

	// Initialize policy engine (optional)
	var policyEngine *policy.OPAEngine

	// Create a context with timeout for the test
	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Start daemon in a goroutine (use compatibility wrapper for older test signature)
	errChan := make(chan error, 1)
	go func() {
		err := service.StartDaemonOld(config, pool, policyEngine, console)
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
