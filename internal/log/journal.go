package logutil

import (
	"log"
	"os"
)

// SetupLogging configures logging for systemd journal or stdout
func SetupLogging() {
	// Check if running under systemd by checking JOURNAL_STREAM environment variable
	if isSystemdService() {
		log.Println("[daemon] Detected systemd environment, logging to journal")
		setupJournalLogging()
	} else {
		log.Println("[daemon] Using standard output logging")
		setupStdoutLogging()
	}
}

// isSystemdService detects if the process is running under systemd
func isSystemdService() bool {
	// Check for JOURNAL_STREAM environment variable
	// This is set by systemd when journald is available
	journalStream := os.Getenv("JOURNAL_STREAM")
	if journalStream != "" {
		return true
	}

	// Additional check for systemd invocation ID
	invocationID := os.Getenv("INVOCATION_ID")
	if invocationID != "" {
		return true
	}

	return false
}

// setupJournalLogging configures logging for systemd journal
func setupJournalLogging() {
	// For systemd journal, we want structured logging without timestamps
	// since journald handles timestamps automatically

	// Set log format to exclude timestamp (journald adds it)
	log.SetFlags(0)

	// Ensure output goes to stdout (journald captures this)
	log.SetOutput(os.Stdout)

	// Note: For more advanced journal integration, we could use
	// github.com/coreos/go-systemd/journal package to send
	// structured messages directly to journald
}

// setupStdoutLogging configures standard logging for non-systemd environments
func setupStdoutLogging() {
	// Standard logging with timestamps for non-systemd environments
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)
}

// LogDaemonEvent logs a daemon lifecycle event with consistent formatting
func LogDaemonEvent(message string) {
	log.Printf("[daemon] %s", message)
}

// LogBootstrapEvent logs a bootstrap event with consistent formatting
func LogBootstrapEvent(message string) {
	log.Printf("[bootstrap] %s", message)
}

// LogPolicyEvent logs a policy bootstrap event with consistent formatting
func LogPolicyEvent(message string) {
	log.Printf("[policy-bootstrap] %s", message)
}

// LogSchemaEvent logs a schema bootstrap event with consistent formatting
func LogSchemaEvent(message string) {
	log.Printf("[schema-bootstrap] %s", message)
}
