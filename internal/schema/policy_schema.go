package schema

type AT1ExampleContext struct {
	Corrupt bool `json:"corrupt"`
}

// AT1CorruptionContext represents attrs expected by dis.policy.at1_corruption
type AT1CorruptionContext struct {
	Corrupt bool `json:"corrupt"`
}
