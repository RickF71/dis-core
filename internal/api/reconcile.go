package api

import "time"

// SchemaWarning represents a version mismatch or other schema import issue
type SchemaWarning struct {
	ID         int        `json:"id"`
	FilePath   string     `json:"file"`
	Domain     string     `json:"domain"`
	SchemaType string     `json:"schema_type"`
	Details    string     `json:"details"`
	Resolved   bool       `json:"resolved"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// DomainWarning represents lineage or linkage issues for domains
type DomainWarning struct {
	ID         int        `json:"id"`
	Domain     string     `json:"domain"`
	Parent     string     `json:"parent"`
	Details    string     `json:"details"`
	Resolved   bool       `json:"resolved"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}
