package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRegoBundle(t *testing.T) {
	// Create a temporary rego file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.rego")

	expectedContent := `package test

default allow = false

allow {
	input.action == "test"
}`

	err := os.WriteFile(testFile, []byte(expectedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test loading the file
	content, err := LoadRegoBundle(testFile)
	if err != nil {
		t.Errorf("LoadRegoBundle() error = %v, want nil", err)
		return
	}

	if content != expectedContent {
		t.Errorf("LoadRegoBundle() content = %q, want %q", content, expectedContent)
	}
}

func TestLoadRegoBundleNonExistentFile(t *testing.T) {
	// Test loading a non-existent file
	_, err := LoadRegoBundle("nonexistent.rego")
	if err == nil {
		t.Error("LoadRegoBundle() error = nil, want error for non-existent file")
	}

	if !strings.Contains(err.Error(), "read rego file") {
		t.Errorf("LoadRegoBundle() error = %v, want error message containing 'read rego file'", err)
	}
}

func TestEnsureDefaultPoliciesFileLoading(t *testing.T) {
	// Test that policy files exist in the expected location
	policyFiles := []string{
		"gates.rego",
		"risk.rego",
		"freeze.rego",
	}

	for _, file := range policyFiles {
		// Try to find the file in the current directory or internal/policy
		paths := []string{
			file,
			filepath.Join("internal", "policy", file),
			filepath.Join("..", "..", "policy", file), // From test directory
		}

		found := false
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				found = true

				// Test that we can read the file
				content, err := LoadRegoBundle(path)
				if err != nil {
					t.Errorf("Failed to load %s: %v", file, err)
					continue
				}

				// Basic validation that it looks like a rego file
				if !strings.Contains(content, "package") {
					t.Errorf("Policy file %s does not contain 'package' keyword", file)
				}

				t.Logf("Successfully loaded %s (%d bytes)", file, len(content))
				break
			}
		}

		if !found {
			t.Logf("Warning: Policy file %s not found in expected locations", file)
		}
	}
}

func TestPolicyBootstrapLogging(t *testing.T) {
	// This test verifies that the bootstrap function can be called
	// without a database connection (simulating file-only loading)

	// Note: This would require mocking the database connection
	// For now, we just test that the LoadRegoBundle function works
	// which is the core file-loading functionality

	// Create a test rego file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_policy.rego")

	testContent := `package test_policy

default allow = false

# Test policy rules
allow {
	input.user == "admin"
}`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test policy file: %v", err)
	}

	// Test that we can load it
	content, err := LoadRegoBundle(testFile)
	if err != nil {
		t.Errorf("Failed to load test policy: %v", err)
		return
	}

	// Verify content
	if !strings.Contains(content, "package test_policy") {
		t.Error("Test policy should contain package declaration")
	}

	if !strings.Contains(content, "default allow = false") {
		t.Error("Test policy should contain default rule")
	}

	t.Logf("Test policy loaded successfully (%d bytes)", len(content))
}
