package schema

import "fmt"

// AuthoritantCollective — authoritant.collective.v1
type AuthoritantCollective struct {
	AuthoritantDefinition `json:",inline" yaml:",inline"`
	Members               []string `json:"members" yaml:"members"`
	Threshold             int      `json:"threshold" yaml:"threshold"`
	Policy                string   `json:"policy,omitempty" yaml:"policy,omitempty"` // majority | unanimous | supermajority
}

func (d AuthoritantCollective) Validate() error {
	if err := d.ValidateBase(); err != nil {
		return err
	}
	if len(d.Members) == 0 {
		return fmt.Errorf("members is required and must not be empty")
	}
	if d.Threshold <= 0 {
		return fmt.Errorf("threshold must be >= 1")
	}
	if d.Threshold > len(d.Members) {
		return fmt.Errorf("threshold (%d) cannot exceed members count (%d)", d.Threshold, len(d.Members))
	}
	if d.Policy != "" {
		if err := oneOf("policy", d.Policy, "majority", "unanimous", "supermajority"); err != nil {
			return err
		}
	}
	return nil
}
