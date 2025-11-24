package envx

import "os"

// InTestBootstrap returns true when the DIS_TEST_BOOTSTRAP environment
// variable is set to "1". This is used by bootstrap and test helpers to
// enable a non-interactive, fast test bootstrap mode. It MUST NOT be used
// to change production behaviour except in tests/CI.
func InTestBootstrap() bool {
	return os.Getenv("DIS_TEST_BOOTSTRAP") == "1"
}

// PolicyAT1Enabled returns true when AT-1 policy enforcement is enabled.
// The env var DIS_POLICY_AT1_ENABLED can be set to "0" or "false" to disable.
func PolicyAT1Enabled() bool {
	v := os.Getenv("DIS_POLICY_AT1_ENABLED")
	if v == "" {
		// Default enabled for dev and tests
		return true
	}
	if v == "0" || v == "false" {
		return false
	}
	return true
}
