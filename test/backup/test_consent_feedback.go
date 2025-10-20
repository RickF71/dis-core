//go:build ignore

package main

import (
	"log"
	"time"

	"dis-core/internal/consent"
	"dis-core/internal/trust"
)

func main() {
	log.Println("🧩 DIS v0.8.8 — Consent Gate and Trust Feedback Test")

	// Simulate USA domain proposing an action
	gate, err := consent.CheckConsent(consent.ConsentInput{
		DomainRef:   "domain.usa",
		ActionRef:   "policy.energy_2025",
		Approvals:   4,
		Rejections:  1,
		MoralWeight: 0.8,
		Quorum:      3,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Apply moral feedback from NOTECH domain
	trust.ApplyFeedback(&trust.TrustFeedback{
		SourceDomain:  "domain.notech",
		TargetDomain:  gate.DomainRef,
		ActionRef:     gate.ActionRef,
		MoralDelta:    0.2, // positive approval
		Confidence:    gate.Confidence,
		SourceReceipt: "mock-receipt-123",
		Timestamp:     time.Now().UTC(),
	})
}
