package domain

import "dis-core/internal/core/domain/spineconfig"

type Dimension int

const (
	Dimension0D Dimension = iota
	DimensionNull
	DimensionAether
	DimensionTerra
	DimensionNuman
	DimensionLima
	DimensionCorporeal
)

func (d Dimension) String() string {
	switch d {
	case Dimension0D:
		return "0d"
	case DimensionNull:
		return "null"
	case DimensionAether:
		return "aether"
	case DimensionTerra:
		return "terra"
	case DimensionNuman:
		return "numan"
	case DimensionLima:
		return "lima"
	case DimensionCorporeal:
		return "corporeal"
	default:
		return "unknown"
	}
}

func IsImmutableDimension(d Dimension) bool {
	return d == Dimension0D
}

func IsValidSpineDimension(d Dimension) bool {
	return d >= Dimension0D && d <= DimensionCorporeal
}

// DimensionFromName maps a canonical domain name to the Dimension value using spineconfig.
func DimensionFromName(name string) (Dimension, bool) {
	if dim, ok := spineconfig.DimensionFor(name); ok {
		return Dimension(dim), true
	}
	return -1, false
}

// ParentNameFor returns the canonical parent name for a child using spineconfig.
func ParentNameFor(child string) (string, bool) {
	return spineconfig.ParentName(child)
}
