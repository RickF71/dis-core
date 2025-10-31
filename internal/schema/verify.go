// AUTOGEN-COPILOT: scaffold verified by RickF71
// Ref: MOAR Phase 2 – Schema Verification
//
// GOAL: Validate integrity of all loaded schema files by comparing
// computed SHA256 hashes and ensuring each contains a "version" field.
// FUNCTIONS: VerifySchemaSet(), summarizeSchemas(), compareVersions()
// OUTPUT: log table showing name, version, and hash of each schema.
// CONTEXT: Called after SchemaVerify() in loader.go to ensure canonical
// schema integrity before route registration.

package schema

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

// VerifySchemaSet takes a list of schema entries and verifies their integrity.
func VerifySchemaSet(schemas []Entry) error {
	for _, entry := range schemas {
		if err := verifySchema(entry); err != nil {
			return fmt.Errorf("schema verification failed for %s: %w", entry.ID, err)
		}
	}
	LogSchemaSummary(schemas)
	return nil
}

// verifySchema checks that the schema has a valid version and that its hash matches the file content.
func verifySchema(entry Entry) error {
	// Check for version field
	if entry.Version == "" {
		return fmt.Errorf("missing version field")
	}

	// Read actual schema file from disk
	data, err := os.ReadFile(entry.Path)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Compute SHA256 of the content
	hash := sha256.Sum256(data)
	computedHash := hex.EncodeToString(hash[:])

	// Compare computed hash to recorded hash
	if computedHash != entry.Hash {
		return fmt.Errorf("hash mismatch for %s: expected %s, got %s", entry.ID, entry.Hash, computedHash)
	}

	log.Printf("✅ Verified schema: ID=%s, Version=%s, Hash=%s", entry.ID, entry.Version, computedHash)
	return nil
}

// summarizeSchemas generates a summary of all schemas for logging purposes.
func summarizeSchemas(schemas []Entry) []map[string]string {
	var summary []map[string]string
	for _, entry := range schemas {
		summary = append(summary, map[string]string{
			"id":      entry.ID,
			"version": entry.Version,
			"hash":    entry.Hash,
		})
	}
	return summary
}

// compareVersions checks if the version of the schema matches the expected format.
func compareVersions(version string) bool {
	// TODO: Add semantic versioning validation if needed
	return version != ""
}

// LogSchemaSummary logs a summary of all verified schemas.
func LogSchemaSummary(schemas []Entry) {
	log.Println("---- Verified Schema Summary ----")
	for _, entry := range schemas {
		log.Printf("ID=%s | Version=%s | Hash=%s", entry.ID, entry.Version, entry.Hash)
	}
	log.Println("---------------------------------")
}
