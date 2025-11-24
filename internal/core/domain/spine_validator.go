package domain

import (
	"context"
	"dis-core/internal/core/domain/spineconfig"
	"fmt"
)

// ValidateSpineTransition verifies that child dimension is exactly one step above parent.
// 0D → null → aether → terra → numan → lima → corporeal
func ValidateSpineTransition(parent, child Dimension) error {
	if !IsValidSpineDimension(parent) || !IsValidSpineDimension(child) {
		return fmt.Errorf("invalid dimension(s)")
	}

	// no one may transition out of 0D except null
	if parent == Dimension0D {
		if child != DimensionNull {
			return fmt.Errorf("0D may only totelevate into null")
		}
		return nil
	}

	// Use canonical spine ordering to validate parent→child adjacency.
	pName := parent.String()
	cName := child.String()
	pIdx, pOk := spineconfig.IndexByName(pName)
	cIdx, cOk := spineconfig.IndexByName(cName)
	if !pOk || !cOk {
		return fmt.Errorf("invalid spine names: %s, %s", pName, cName)
	}
	if cIdx != pIdx+1 {
		return fmt.Errorf("invalid spine transition: %s → %s", pName, cName)
	}

	return nil
}

// InferParentDimension attempts to infer a parent's Dimension by reading its name or domain_type
// from the provided query function. The function should query the DB and return (domain_type, name, found).
// This helper is provided to keep DB access co-located with validation logic.
func InferParentDimension(ctx context.Context, queryParent func(ctx context.Context) (string, string, bool)) (Dimension, error) {
	domainType, name, found := queryParent(ctx)
	if !found {
		return -1, fmt.Errorf("parent domain not found")
	}
	// Prefer domain_type when available
	switch domainType {
	case "corporeal":
		return DimensionCorporeal, nil
	case "lima":
		return DimensionLima, nil
	case "numan", "numen":
		return DimensionNuman, nil
	case "terra":
		return DimensionTerra, nil
	case "aether":
		return DimensionAether, nil
	case "null":
		return DimensionNull, nil
	}
	// Prefer canonical spine mapping where possible
	if dim, ok := DimensionFromName(name); ok {
		return dim, nil
	}
	// legacy fallbacks
	if name == "domain.null" {
		return DimensionNull, nil
	}
	if name == "domain.aether" {
		return DimensionAether, nil
	}
	if name == "domain.terra" {
		return DimensionTerra, nil
	}
	if name == "domain.numan" || name == "domain.numen" {
		return DimensionNuman, nil
	}
	if name == "domain.lima" {
		return DimensionLima, nil
	}
	if name == "domain.corporeal" {
		return DimensionCorporeal, nil
	}
	return -1, fmt.Errorf("unable to infer parent dimension from type=%q name=%q", domainType, name)
}
