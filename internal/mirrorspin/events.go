package mirrorspin

// MirrorEvent represents a reflection or counterspin event
type MirrorEvent struct {
    Type        string
    Source      string
    Target      string
    EntityType  string
    EntityID    string
    Timestamp   string
    Hash        string
    DiffSummary string
}
