package policy

// EngineConfig: file paths + (future) plugs for state/authz.
type EngineConfig struct {
	PathFreezeRego     string
	PathGatesRego      string
	PathRiskRego       string
	PathThresholdsJSON string
	PathCIRulesJSON    string
	PathRedactionYAML  string
	PathCedarSchema    string
	StateProvider      interface{}
	AuthZ              interface{}
}
