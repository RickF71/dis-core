package types

// LifePushSubstrate defines a substrate layer with formal coherence thresholds.
type LifePushSubstrate struct {
	ID                  string  `json:"id" db:"id"`
	Layer               string  `json:"layer" db:"layer"` // LifePush, Terra, Limen
	CoherenceThreshold  float64 `json:"coherence_threshold" db:"coherence_threshold"`
	EnergyFlowMin       float64 `json:"energy_flow_min" db:"energy_flow_min"`
	ConsentIntegrityMin float64 `json:"consent_integrity_min" db:"consent_integrity_min"`
	SuccessorLayer      string  `json:"successor_layer,omitempty" db:"successor_layer"`
	ObserverDomain      string  `json:"observer_domain,omitempty" db:"observer_domain"`
	Active              bool    `json:"active" db:"active"`
}
