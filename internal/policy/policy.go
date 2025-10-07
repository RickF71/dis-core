package policy

import (
	"fmt"
	"os"
	"sort"

	"dis-core/internal/util"

	"gopkg.in/yaml.v3"
)

type Policy struct {
	Allowed map[string][]string `yaml:"allowed"` // domain -> scopes
	Deny    []string            `yaml:"deny"`
}

func Load(path string) (*Policy, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	var p Policy
	if err := yaml.Unmarshal(b, &p); err != nil {
		return nil, "", err
	}
	// Normalize (sort for stable checksum/printing)
	for k := range p.Allowed {
		s := p.Allowed[k]
		sort.Strings(s)
		p.Allowed[k] = s
	}
	sort.Strings(p.Deny)
	sum := util.ChecksumHex(b)
	return &p, sum, nil
}

func (p *Policy) IsDomainDenied(domain string) bool {
	for _, d := range p.Deny {
		if d == domain {
			return true
		}
	}
	return false
}

func (p *Policy) IsAllowed(domain, scope string) bool {
	list, ok := p.Allowed[domain]
	if !ok {
		return false
	}
	for _, s := range list {
		if s == scope {
			return true
		}
	}
	return false
}

func PrintSummary(p *Policy) {
	fmt.Println("deny:", p.Deny)
	fmt.Println("allowed:")
	for d, scopes := range p.Allowed {
		fmt.Printf("  %s:\n", d)
		for _, s := range scopes {
			fmt.Printf("    - %s\n", s)
		}
	}
}
