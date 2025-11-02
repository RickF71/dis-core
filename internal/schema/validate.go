package schema

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ISO8601 validates an RFC3339 timestamp (a common ISO-8601 profile).
func ISO8601(ts string) error {
	if ts == "" {
		return errors.New("timestamp is empty")
	}
	if _, err := time.Parse(time.RFC3339, ts); err != nil {
		return fmt.Errorf("invalid timestamp %q: %w", ts, err)
	}
	return nil
}

func requireNonEmpty(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}

func oneOf(field, value string, allowed ...string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return fmt.Errorf("%s=%q not in %v", field, value, allowed)
}
