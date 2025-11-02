package importer

import "fmt"

// ValidateBasic ensures type/version/domain fields exist
func ValidateBasic(obj map[string]any) error {
	req := []string{"type", "version", "domain"}
	for _, key := range req {
		if obj[key] == nil || obj[key] == "" {
			return fmt.Errorf("missing required field: %s", key)
		}
	}
	return nil
}
