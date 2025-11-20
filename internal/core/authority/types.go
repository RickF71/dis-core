package authority

import (
	"time"

	"github.com/google/uuid"
)

// AuthorityStatus is the canonical status payload used by the
// core authority engine and API delegators. This file is part of
// MX-3.1 status migration.
type AuthorityStatus struct {
	EngineHealthy bool   `json:"engine_healthy"`
	FreezeActive  bool   `json:"freeze_active"`
	Notes         string `json:"notes,omitempty"`
}

// IntrospectInfo is the minimal introspection payload introduced in MX-3.2.
type IntrospectInfo struct {
	EngineVersion string `json:"engine_version"`
	Notes         string `json:"notes,omitempty"`
}

// LineageNode is a minimal node used for lineage/ancestry responses (MX-3.3).
type LineageNode struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	ParentID  string `json:"parent_id,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Notes     string `json:"notes,omitempty"`
}

// LineageResult is a small wrapper describing a target's ancestry and children.
type LineageResult struct {
	Root     *LineageNode   `json:"root"`
	Children []*LineageNode `json:"children"`
	Notes    string         `json:"notes,omitempty"`
}

// BranchInfo is a minimal struct describing branch metadata for a domain.
type BranchInfo struct {
	DomainID    uuid.UUID `json:"domain_id,omitempty"`
	ParentID    uuid.UUID `json:"parent_id,omitempty"`
	BranchDepth int       `json:"branch_depth,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	Notes       string    `json:"notes,omitempty"`
}

// CreateBranchRequest is the engine-level request for creating a branch.
type CreateBranchRequest struct {
	OriginalDomainID uuid.UUID
	NewParentID      uuid.UUID
	BranchName       string
	Reason           string
	CreatedBy        uuid.UUID
	Metadata         map[string]interface{}
}

// OverrideResult is a minimal result returned by the Engine.Override operation.
type OverrideResult struct {
	DomainID string                 `json:"domain_id,omitempty"`
	Changed  map[string]interface{} `json:"changed,omitempty"`
	Notes    string                 `json:"notes,omitempty"`
}

// Provide a compatibility alias for older code referencing AuthorityEngine
// (many callers still refer to AuthorityEngine); this is a harmless alias
// so both naming conventions work during migration.
type AuthorityEngine = Engine
