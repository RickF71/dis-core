package domain

import "testing"

func TestDimensionString(t *testing.T) {
	if DimensionAether.String() != "aether" {
		t.Fatalf("expected 'aether', got %s", DimensionAether.String())
	}
	if DimensionNull.String() != "null" {
		t.Fatalf("expected 'null', got %s", DimensionNull.String())
	}
}

func TestIsImmutableDimension(t *testing.T) {
	if !IsImmutableDimension(Dimension0D) {
		t.Fatalf("0D must be immutable")
	}
	if IsImmutableDimension(DimensionNull) {
		t.Fatalf("only 0D is immutable; null is the first policy layer")
	}
}

func TestIsValidSpineDimension(t *testing.T) {
	if !IsValidSpineDimension(DimensionCorporeal) {
		t.Fatalf("corporeal must be valid spine dimension")
	}
	if IsValidSpineDimension(Dimension(99)) {
		t.Fatalf("invalid dimension accepted as valid")
	}
}
