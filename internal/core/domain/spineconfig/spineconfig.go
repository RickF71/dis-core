package spineconfig

// SpineEntry represents an ordered entry in the canonical domain spine.
type SpineEntry struct {
	Name      string
	Dimension int
}

// CanonicalSpine returns the ordered domain spine used across the codebase.
func CanonicalSpine() []SpineEntry {
	return []SpineEntry{
		{Name: "void", Dimension: 0},
		{Name: "null", Dimension: 1},
		{Name: "aether", Dimension: 2},
		{Name: "terra", Dimension: 3},
		{Name: "numen", Dimension: 4},
		{Name: "lima", Dimension: 5},
		{Name: "corporeal", Dimension: 6},
	}
}

// IndexByName returns the index (0-based) of the named spine entry and whether it was found.
func IndexByName(name string) (int, bool) {
	// normalize common variants
	n := name
	if len(n) > 7 && n[:7] == "domain." {
		n = n[7:]
	}
	if n == "numan" {
		n = "numen"
	}
	for i, e := range CanonicalSpine() {
		if e.Name == n {
			return i, true
		}
	}
	return -1, false
}

// ParentName returns the canonical parent name for the given child name.
// It returns ok=false if the child is not found or has no parent.
func ParentName(child string) (string, bool) {
	idx, ok := IndexByName(child)
	if !ok || idx == 0 {
		return "", false
	}
	return CanonicalSpine()[idx-1].Name, true
}

// DimensionFor returns the numeric dimension for a canonical name.
func DimensionFor(name string) (int, bool) {
	for _, e := range CanonicalSpine() {
		if e.Name == name {
			return e.Dimension, true
		}
	}
	return -1, false
}
