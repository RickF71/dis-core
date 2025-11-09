package bootstrap

import (
	"log"

	"dis-core/internal/policy"
)

// PolicyComponents holds the initialized policy engine components
type PolicyComponents struct {
	Engine policy.PolicyEngine
}

// InitializePolicyEngine sets up the policy engine with local rego modules
func InitializePolicyEngine() (*PolicyComponents, error) {
	log.Println("🔒 Initializing policy engine...")

	modules, err := policy.LoadLocalModules()
	if err != nil {
		return nil, err
	}

	engine, err := policy.NewEngine(modules)
	if err != nil {
		return nil, err
	}

	log.Println("✅ Policy engine ready with local modules")

	return &PolicyComponents{
		Engine: engine,
	}, nil
}
