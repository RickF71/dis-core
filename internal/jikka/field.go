package jikka

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"dis-core/internal/ledger"
)

// JikkaField represents an emergent moral field between two domains.
// It can exist only where both domains act from free will and mutual recognition.
type JikkaField struct {
	ID            string    `json:"id"`
	DomainA       string    `json:"domain_a"`
	DomainB       string    `json:"domain_b"`
	FreeWillA     bool      `json:"free_will_a"`
	FreeWillB     bool      `json:"free_will_b"`
	Recognition   bool      `json:"recognition"`
	Consent       bool      `json:"consent"`
	Reflection    bool      `json:"reflection"`
	FieldStrength float64   `json:"field_strength"` // 0.0 – 1.0 coherence
	Integrity     string    `json:"integrity"`      // absent, latent, active, transcendent
	CreatedAt     time.Time `json:"created_at"`
	DissolvedAt   time.Time `json:"dissolved_at,omitempty"`
}

// Validate determines if a Jikka field can exist between two domains.
func (j *JikkaField) Validate() error {
	if !j.FreeWillA || !j.FreeWillB {
		j.Integrity = "absent"
		return errors.New("❌ Jikka cannot form: missing free will on one or both sides")
	}
	if !j.Recognition {
		j.Integrity = "latent"
		return errors.New("⚠️ Jikka latent: recognition missing")
	}
	if !j.Consent {
		j.Integrity = "latent"
		return errors.New("⚠️ Jikka latent: consent missing")
	}
	if !j.Reflection {
		j.Integrity = "latent"
		return errors.New("⚠️ Jikka latent: reflection missing")
	}

	j.Integrity = "active"
	return nil
}

// EvaluateFieldStrength computes a normalized relational coherence score.
func (j *JikkaField) EvaluateFieldStrength() {
	// Basic heuristic: count how many of the core conditions are met
	total := 4.0 // recognition, consent, reflection, free will parity
	score := 0.0
	if j.FreeWillA && j.FreeWillB {
		score += 1
	}
	if j.Recognition {
		score += 1
	}
	if j.Consent {
		score += 1
	}
	if j.Reflection {
		score += 1
	}
	j.FieldStrength = score / total

	if j.FieldStrength == 1.0 {
		j.Integrity = "transcendent"
	}
}

// CreateJikka attempts to form a moral connection between two domains.
// Returns a JikkaField instance and records a receipt.
func CreateJikka(domainA, domainB string, freeWillA, freeWillB, recognition, consent, reflection bool) (*JikkaField, error) {
	field := &JikkaField{
		ID:          ledger.GenerateUUID(),
		DomainA:     domainA,
		DomainB:     domainB,
		FreeWillA:   freeWillA,
		FreeWillB:   freeWillB,
		Recognition: recognition,
		Consent:     consent,
		Reflection:  reflection,
		CreatedAt:   time.Now().UTC(),
	}

	if err := field.Validate(); err != nil {
		log.Println(err)
	} else {
		field.EvaluateFieldStrength()
	}

	data, _ := json.MarshalIndent(field, "", "  ")
	log.Printf("🌌 Jikka Field created between %s ↔ %s:\n%s", domainA, domainB, string(data))

	// Save receipt for provenance
	ledger.SaveReceipt(&ledger.Receipt{
		Action:     "jikka.create",
		By:         domainA,
		ConsentRef: "",
		Status:     field.Integrity,
		Comments:   "Jikka field evaluation between domains",
	})

	return field, nil
}

// Dissolve ends a Jikka field gracefully.
func (j *JikkaField) Dissolve(reason string) {
	j.DissolvedAt = time.Now().UTC()
	j.Integrity = "absent"

	// Serialize for audit log
	data, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		log.Printf("🌑 Failed to serialize Jikka field during dissolve: %v", err)
	} else {
		log.Printf("🌑 Jikka Field dissolved [%s]: %s\n%s", j.ID, reason, string(data))
	}

	// Record dissolution as a receipt
	ledger.SaveReceipt(&ledger.Receipt{
		Action:   "jikka.dissolve",
		By:       j.DomainA,
		Status:   "ended",
		Comments: reason,
	})
}
