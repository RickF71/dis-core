package consent

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ---- Public Types ----

// ActorRef represents an initiator/affected party at decision time.
// Trust/Legitimacy may be hydrated from your identity/ledger layer.
type ActorRef struct {
	ID         string  `json:"id"`
	Domain     string  `json:"domain"`
	Trust      float64 `json:"trust"`      // 0.0 - 1.0
	Legitimacy float64 `json:"legitimacy"` // 0.0 - 1.0 (domain or seat-derived)
}

// ConsentRequest is the input to the Gate.
type ConsentRequest struct {
	Action    string            `json:"action"`     // e.g., "trade.execute", "policy.update"
	SchemaRef string            `json:"schema_ref"` // receipt/schema binding
	Initiator ActorRef          `json:"initiator"`
	Affected  []ActorRef        `json:"affected"`
	Metadata  map[string]string `json:"metadata,omitempty"` // free-form; kept in receipt
	Context   map[string]any    `json:"context,omitempty"`  // domain-specific
}

// Decision captures the gate outcome and moral math used.
type Decision struct {
	Allowed        bool       `json:"allowed"`
	Reason         string     `json:"reason"`
	Legitimacy     float64    `json:"legitimacy"`
	ThrottleUntil  *time.Time `json:"throttle_until,omitempty"`
	AppliedRules   []string   `json:"applied_rules"`
	TrustDelta     float64    `json:"trust_delta"`
	EthicsDelta    float64    `json:"ethics_delta"`
	LegitimacyRule string     `json:"legitimacy_rule"` // the decisive rule id
}

// Receipt is a minimal shape the Gate can emit.
// You’ll likely replace this with your internal/ledger type.
type Receipt struct {
	ID             string            `json:"id"`
	Time           time.Time         `json:"time"`
	Action         string            `json:"action"`
	SchemaRef      string            `json:"schema_ref"`
	InitiatorID    string            `json:"initiator_id"`
	AffectedIDs    []string          `json:"affected_ids"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	Decision       string            `json:"decision"` // "allowed"|"blocked"|"throttled"
	TrustChange    float64           `json:"trust_change"`
	EthicsChange   float64           `json:"ethics_change"`
	LegitimacyRef  string            `json:"legitimacy_ref"` // rule id/source
	ConsentVersion string            `json:"consent_version"`
}

// FeedbackSink receives receipts to drive moral feedback loops.
type FeedbackSink interface {
	Apply(ctx context.Context, rcpt Receipt) error
}

// ReceiptPoster lets you plug your ledger post operation.
type ReceiptPoster interface {
	Post(ctx context.Context, rcpt Receipt) (string, error) // returns receipt ID
}

// ---- Config (loaded from YAML) ----

// Config matches dis_consent.v0.1.yaml (subset).
type Config struct {
	Version   string       `json:"version" yaml:"version"`
	GateRules []GateRule   `json:"gate_rules" yaml:"gate_rules"`
	Weights   Weights      `json:"weights" yaml:"weights"`
	Throttle  ThrottleRule `json:"throttle" yaml:"throttle"`
}

type GateRule struct {
	ID          string  `json:"id" yaml:"id"`
	Threshold   float64 `json:"threshold" yaml:"threshold"` // 0..1 (use -1 for "dynamic" then resolve at runtime)
	Description string  `json:"description" yaml:"description"`
}

type Weights struct {
	TrustIncrease float64 `json:"trust_increase" yaml:"trust_increase"`
	TrustDecrease float64 `json:"trust_decrease" yaml:"trust_decrease"`
	EthicsBonus   float64 `json:"ethics_bonus" yaml:"ethics_bonus"`
	EthicsPenalty float64 `json:"ethics_penalty" yaml:"ethics_penalty"`
}

type ThrottleRule struct {
	Enabled       bool          `json:"enabled" yaml:"enabled"`
	LowTrustFloor float64       `json:"low_trust_floor" yaml:"low_trust_floor"` // e.g., 0.3
	Backoff       time.Duration `json:"-" yaml:"backoff_ms"`                    // decode as ms in loader
}

// ---- Gate ----

type Gate struct {
	mu      sync.RWMutex
	cfg     *Config
	sink    FeedbackSink
	poster  ReceiptPoster
	timeNow func() time.Time
	version string
}

// NewGate constructs a Gate with injected sinks/posters (can be nil during bootstrapping).
func NewGate(cfg *Config, version string, sink FeedbackSink, poster ReceiptPoster) *Gate {
	return &Gate{
		cfg:     cfg,
		sink:    sink,
		poster:  poster,
		timeNow: time.Now,
		version: version,
	}
}

// VerifyConsent checks legitimacy/consent without posting a receipt.
func (g *Gate) VerifyConsent(ctx context.Context, req ConsentRequest) (Decision, error) {
	if g.cfg == nil {
		return Decision{}, errors.New("consent gate not configured")
	}

	// Base legitimacy: initiator legitimacy blended with transparency reciprocity.
	// You can make this richer (e.g., weighted by affected party count).
	leg := clamp((req.Initiator.Legitimacy+req.Initiator.Trust)/2.0, 0, 1)

	applied := []string{}
	decisive := "personal_autonomy"

	// Apply rules in priority order (as listed)
	for _, r := range g.cfg.GateRules {
		applied = append(applied, r.ID)
		switch r.ID {
		case "personal_autonomy":
			// Require explicit consent when individuals are affected
			// Here we simulate "explicit consent present?" via Metadata["consent"]="yes"
			if anyIndividuals(req) && (req.Metadata == nil || req.Metadata["consent"] != "yes") {
				return Decision{
					Allowed:        false,
					Reason:         "missing explicit consent",
					Legitimacy:     leg,
					AppliedRules:   applied,
					TrustDelta:     g.cfg.Weights.TrustDecrease,
					EthicsDelta:    g.cfg.Weights.EthicsPenalty,
					LegitimacyRule: r.ID,
				}, nil
			}
			if leg < r.Threshold {
				return Decision{
					Allowed:        false,
					Reason:         "insufficient legitimacy for personal autonomy",
					Legitimacy:     leg,
					AppliedRules:   applied,
					TrustDelta:     g.cfg.Weights.TrustDecrease,
					EthicsDelta:    g.cfg.Weights.EthicsPenalty,
					LegitimacyRule: r.ID,
				}, nil
			}

		case "reciprocal_transparency":
			// If all affected parties are informed, boost perceived legitimacy slightly
			if informedAll(req) {
				leg = clamp(leg+0.03, 0, 1)
			}
			decisive = r.ID

		case "trust_decay":
			// Declarative hook; actual decay is applied by feedback loop from prior receipts
			decisive = r.ID

		case "moral_feedback":
			// Declarative hook; actual post-decision effects happen in feedback loop
			decisive = r.ID
		}
	}

	// Optional throttling: low trust leads to timed backoff instead of hard block
	if g.cfg.Throttle.Enabled && req.Initiator.Trust < g.cfg.Throttle.LowTrustFloor {
		tu := g.timeNow().Add(g.cfg.Throttle.Backoff)
		return Decision{
			Allowed:        false,
			Reason:         "throttled: low trust backoff",
			Legitimacy:     leg,
			ThrottleUntil:  &tu,
			AppliedRules:   applied,
			TrustDelta:     g.cfg.Weights.TrustDecrease,
			EthicsDelta:    0,
			LegitimacyRule: "throttle.low_trust",
		}, nil
	}

	// Allowed path: modest bonuses for clean consent path
	return Decision{
		Allowed:        true,
		Reason:         "consent verified",
		Legitimacy:     leg,
		AppliedRules:   applied,
		TrustDelta:     g.cfg.Weights.TrustIncrease,
		EthicsDelta:    g.cfg.Weights.EthicsBonus,
		LegitimacyRule: decisive,
	}, nil
}

// AuthorizeAction runs VerifyConsent, emits a receipt, and forwards to feedback sink.
func (g *Gate) AuthorizeAction(ctx context.Context, req ConsentRequest) (Decision, *Receipt, error) {
	dec, err := g.VerifyConsent(ctx, req)
	if err != nil {
		return Decision{}, nil, err
	}

	rcpt := &Receipt{
		ID:             "", // filled by poster if available
		Time:           g.timeNow(),
		Action:         req.Action,
		SchemaRef:      req.SchemaRef,
		InitiatorID:    req.Initiator.ID,
		AffectedIDs:    collectIDs(req.Affected),
		Metadata:       req.Metadata,
		Decision:       decisionWord(dec),
		TrustChange:    dec.TrustDelta,
		EthicsChange:   dec.EthicsDelta,
		LegitimacyRef:  dec.LegitimacyRule,
		ConsentVersion: g.version,
	}

	// Persist receipt if a poster is configured.
	if g.poster != nil {
		id, perr := g.poster.Post(ctx, *rcpt)
		if perr != nil {
			return dec, rcpt, perr
		}
		rcpt.ID = id
	}

	// Feed moral dynamics.
	if g.sink != nil {
		_ = g.sink.Apply(ctx, *rcpt) // non-fatal; log in real impl
	}

	return dec, rcpt, nil
}

// ---- Helpers ----

func anyIndividuals(req ConsentRequest) bool {
	// Placeholder heuristic: if any affected has Domain "persona" or empty -> treat as individual.
	for _, a := range req.Affected {
		if a.Domain == "" || a.Domain == "persona" {
			return true
		}
	}
	return false
}

func informedAll(req ConsentRequest) bool {
	// Placeholder: treat Metadata["informed_all"]="yes" as proof of reciprocal transparency.
	return req.Metadata != nil && req.Metadata["informed_all"] == "yes"
}

func collectIDs(xs []ActorRef) []string {
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		out = append(out, x.ID)
	}
	return out
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func decisionWord(d Decision) string {
	switch {
	case d.Allowed:
		return "allowed"
	case d.ThrottleUntil != nil:
		return "throttled"
	default:
		return "blocked"
	}
}
