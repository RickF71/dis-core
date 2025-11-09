package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCoreSchema(t *testing.T) {
	// Test that LoadCoreSchema can run without errors
	registry, err := LoadCoreSchema()
	if err != nil {
		t.Errorf("LoadCoreSchema() error = %v, want nil", err)
		return
	}

	if registry == nil {
		t.Error("LoadCoreSchema() returned nil registry")
		return
	}

	t.Logf("LoadCoreSchema() completed successfully")
}

func TestLoadSchemaFile(t *testing.T) {
	// Create a temporary schema file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_schema.yaml")

	testContent := `---
meta:
  schema_id: "test.schema"
  schema_version: "v1.0"
  description: "Test schema for unit testing"

properties:
  name:
    type: string
    description: "Test property"
  value:
    type: number
    description: "Test numeric value"
`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test schema file: %v", err)
	}

	// Test loading the schema file
	registry := NewRegistry()
	err = loadSchemaFile(registry, testFile)
	if err != nil {
		t.Errorf("loadSchemaFile() error = %v, want nil", err)
		return
	}

	t.Logf("Successfully loaded test schema file")
}

func TestLoadSchemaFileNonExistent(t *testing.T) {
	// Test loading a non-existent file
	registry := NewRegistry()
	err := loadSchemaFile(registry, "nonexistent_schema.yaml")
	if err == nil {
		t.Error("loadSchemaFile() error = nil, want error for non-existent file")
		return
	}

	if !strings.Contains(err.Error(), "read file") {
		t.Errorf("loadSchemaFile() error = %v, want error message containing 'read file'", err)
	}
}

func TestSchemaFileDetection(t *testing.T) {
	// Test that schema files exist in expected locations
	schemaDirs := []string{
		"internal/schema",
		"schemas",
		".", // Current directory
	}

	foundSchemas := 0

	for _, dir := range schemaDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".yaml" || ext == ".yml" || ext == ".json" {
				// Try to read the file
				data, err := os.ReadFile(path)
				if err != nil {
					t.Logf("Warning: Could not read schema file %s: %v", path, err)
					return nil
				}

				foundSchemas++
				t.Logf("Found schema file: %s (%.1f KB)", path, float64(len(data))/1024.0)
			}

			return nil
		})

		if err != nil {
			t.Logf("Warning: Error scanning directory %s: %v", dir, err)
		}
	}

	if foundSchemas == 0 {
		t.Log("Warning: No schema files found in expected directories")
	} else {
		t.Logf("Found %d potential schema files", foundSchemas)
	}
}

func TestRegistryCreation(t *testing.T) {
	// Test that we can create a new registry
	registry := NewRegistry()
	if registry == nil {
		t.Error("NewRegistry() returned nil")
		return
	}

	// Test ByKey method
	entries := registry.ByKey()
	if entries == nil {
		t.Error("Registry.ByKey() returned nil")
		return
	}

	t.Logf("Registry created successfully with %d entries", len(entries))
}
