package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateDomainUID_TableDriven(t *testing.T) {
	cases := []struct {
		name  string
		input string
		ok    bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", true},
		{"empty string", "", false},
		{"not a uuid", "not-a-uuid", false},
		{"short hex", "1234", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateDomainUID(c.input)
			if c.ok {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
