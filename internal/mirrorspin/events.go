package mirrorspin

// MirrorEvent represents a reflection or counterspin event
type MirrorEvent struct {
	Type        string
	Source      string
	Target      string
	EntityType  string
	EntityID    string
	CreatedAt   string
	Hash        string
	DiffSummary string
}
