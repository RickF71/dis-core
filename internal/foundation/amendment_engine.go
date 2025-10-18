package foundation

import (
	"errors"
	"fmt"
	"time"

	"dis-core/internal/registry"
)

type AmendmentEvent struct {
	AmendmentID       string
	TargetPath        string   // relative or absolute path to canon file
	NewDefinitionBlob []byte   // the fully rendered new YAML
	RatifiedBy        []string // domain IDs
	ConsensusLevel    float64
	EffectiveAt       time.Time
}

func ApplyAmendment(repoRoot string, event AmendmentEvent) error {
	if event.ConsensusLevel < 0.75 {
		return errors.New("consensus threshold not met (>= 0.75 required)")
	}
	if err := registry.ReplaceByPath(event.TargetPath, event.NewDefinitionBlob, repoRoot); err != nil {
		return fmt.Errorf("amendment failed: %w", err)
	}
	// TODO receipts.Record("dis.event.amendment", event.AmendmentID, event.TargetPath)
	return nil
}
