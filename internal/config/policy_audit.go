package config

import (
	"os"
	"strconv"
)

// GOV-4: Policy and audit configuration

// PolicyAuditConfig holds configuration for policy evaluation and audit logging
type PolicyAuditConfig struct {
	// Policy settings
	PolicyEnabled   bool   // DIS_POLICY_ENABLED (default: true)
	PolicyLoad      string // DIS_POLICY_LOAD: fs|db (default: fs)
	PolicyBundleDir string // DIS_POLICY_BUNDLE_DIR (default: ./policies)

	// Audit settings
	AuditEnabled        bool   // DIS_AUDIT_ENABLED (default: true)
	AuditReceiptKind    string // DIS_AUDIT_RECEIPT_KIND (default: ci.call.v1)
	AuditDecisionsTable string // DIS_DECISIONS_TABLE (default: authority_decisions)
}

// ParsePolicyAuditConfig reads configuration from environment variables
func ParsePolicyAuditConfig() PolicyAuditConfig {
	return PolicyAuditConfig{
		PolicyEnabled:   parseBool("DIS_POLICY_ENABLED", true),
		PolicyLoad:      parseString("DIS_POLICY_LOAD", "fs"),
		PolicyBundleDir: parseString("DIS_POLICY_BUNDLE_DIR", "./policies"),

		AuditEnabled:        parseBool("DIS_AUDIT_ENABLED", true),
		AuditReceiptKind:    parseString("DIS_AUDIT_RECEIPT_KIND", "ci.call.v1"),
		AuditDecisionsTable: parseString("DIS_DECISIONS_TABLE", "authority_decisions"),
	}
}

// parseBool reads a boolean from environment with default fallback
func parseBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}

	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	return parsed
}

// parseString reads a string from environment with default fallback
func parseString(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

// Validate checks if the configuration is valid
func (c *PolicyAuditConfig) Validate() error {
	// Add validation logic if needed
	return nil
}

// IsDevelopmentMode returns true if both policy and audit are disabled
func (c *PolicyAuditConfig) IsDevelopmentMode() bool {
	return !c.PolicyEnabled && !c.AuditEnabled
}
