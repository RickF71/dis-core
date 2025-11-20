package domain

// MX-1 placeholder for domain structures.
// Later: null/void/lima/corporeal model, domain tree logic, registry, lookup, invariants.
type Domain struct {
	ID   string
	Name string
	Type string // "null", "void", "lima", "corporeal", etc.
}
