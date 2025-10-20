package canon

// CanonicalSchema represents a minimal self-describing domain entry
// that can be exported or re-imported in YAML form.
type CanonicalSchema struct {
	ID          string  `yaml:"id"`
	Name        string  `yaml:"name"`
	ParentID    *string `yaml:"parent_id,omitempty"`
	IsNotech    bool    `yaml:"is_notech"`
	InsideReq   bool    `yaml:"requires_inside_domain"`
	LastUpdated string  `yaml:"updated_at"`
}
