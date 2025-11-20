package authority

import "context"

// GetLineage returns a structured lineage result for the target ID.
// MX-3.3: simple placeholder; full ancestry logic will be migrated in MX-3.4.
func (e *Engine) GetLineage(ctx context.Context, targetID string) (*LineageResult, error) {
	// MX-3.3: simple placeholder, real ancestry logic arrives in MX-3.4
	result := &LineageResult{
		Root: &LineageNode{
			ID:   targetID,
			Type: "placeholder",
			Notes: "MX-3.3: lineage structure stub. " +
				"Real ancestry resolution will be migrated in MX-3.4.",
		},
		Children: []*LineageNode{},
		Notes:    "MX-3.3 lineage stub",
	}
	return result, nil
}
