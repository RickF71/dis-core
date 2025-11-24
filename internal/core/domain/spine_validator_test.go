package domain

import "testing"

func TestValidSpineTransitions(t *testing.T) {
	table := []struct {
		parent Dimension
		child  Dimension
	}{
		{Dimension0D, DimensionNull},
		{DimensionNull, DimensionAether},
		{DimensionAether, DimensionTerra},
		{DimensionTerra, DimensionNuman},
		{DimensionNuman, DimensionLima},
		{DimensionLima, DimensionCorporeal},
	}

	for _, tc := range table {
		if err := ValidateSpineTransition(tc.parent, tc.child); err != nil {
			t.Fatalf("expected valid transition %s → %s: %v", tc.parent.String(), tc.child.String(), err)
		}
	}
}

func TestInvalidSpineTransitions(t *testing.T) {
	table := []struct {
		parent Dimension
		child  Dimension
	}{
		{Dimension0D, DimensionAether},
		{DimensionAether, DimensionCorporeal},
		{DimensionCorporeal, Dimension0D},
		{DimensionTerra, DimensionNull},
	}

	for _, tc := range table {
		if err := ValidateSpineTransition(tc.parent, tc.child); err == nil {
			t.Fatalf("expected failure for invalid transition %s → %s", tc.parent.String(), tc.child.String())
		}
	}
}
