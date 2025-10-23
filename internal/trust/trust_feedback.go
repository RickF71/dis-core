package trust

import (
	"encoding/json"
	"log"
	"math"
	"time"

	"dis-core/internal/ledger"
)

type TrustFeedback struct {
	FeedbackID    string    `json:"feedback_id"`
	SourceDomain  string    `json:"source_domain"`
	TargetDomain  string    `json:"target_domain"`
	ActionRef     string    `json:"action_ref"`
	MoralDelta    float64   `json:"moral_delta"` // -1.0 → +1.0
	Confidence    float64   `json:"confidence"`  // 0.0–1.0
	TrustScore    float64   `json:"trust_score"` // evolving trust
	CreatedAt     time.Time `json:"created_at"`
	SourceReceipt string    `json:"source_receipt"`
}

// TrustMap simulates dynamic trust evolution across domains.
var TrustMap = map[string]float64{}

// ApplyFeedback updates a target domain’s trust score
// using a decaying memory of prior trust.
func ApplyFeedback(fb *TrustFeedback) {
	prev := TrustMap[fb.TargetDomain]
	newScore := prev*0.9 + fb.MoralDelta*fb.Confidence*0.1
	newScore = math.Max(0, math.Min(1, newScore)) // clamp 0–1
	fb.TrustScore = math.Round(newScore*100) / 100

	TrustMap[fb.TargetDomain] = fb.TrustScore
	fb.CreatedAt = time.Now().UTC()
	fb.FeedbackID = ledger.GenerateUUID()

	data, _ := json.MarshalIndent(fb, "", "  ")
	log.Printf("🤝 TrustFeedback applied:\n%s", string(data))
}
