package envx

import "os"

// InTestBootstrap returns true when the DIS_TEST_BOOTSTRAP environment
// variable is set to "1". This is used by bootstrap and test helpers to
// enable a non-interactive, fast test bootstrap mode. It MUST NOT be used
// to change production behaviour except in tests/CI.
func InTestBootstrap() bool {
return os.Getenv("DIS_TEST_BOOTSTRAP") == "1"
}
