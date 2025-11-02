package schema

import "fmt"

// AuthoritantSingle — authoritant.single.v1
type AuthoritantSingle struct {
	AuthoritantDefinition `json:",inline" yaml:",inline"`
	Holder                string   `json:"holder" yaml:"holder"`
	Keys                  []string `json:"keys" yaml:"keys"`
}

func (d AuthoritantSingle) Validate() error {
	if err := d.ValidateBase(); err != nil {
		return err
	}
	if err := requireNonEmpty("holder", d.Holder); err != nil {
		return err
	}
	if len(d.Keys) == 0 {
		return fmt.Errorf("keys is required and must contain at least one entry")
	}
	return nil
}
