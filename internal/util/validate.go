package util

import (
	"fmt"

	"github.com/google/uuid"
)

// ValidateDomainUID returns nil if the provided input is a valid UUID string.
// It returns an error describing the problem otherwise.
func ValidateDomainUID(id string) error {
	if id == "" {
		return fmt.Errorf("domain id is required")
	}
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid domain id: %w", err)
	}
	return nil
}
