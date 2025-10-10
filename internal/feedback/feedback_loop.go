package feedback

import (
	"context"
	"sync"
	"time"
)

// Minimal mirror of consent.Receipt for decoupling.
// Replace with your internal ledger receipt type if desired.
type Receipt struct {
	ID            string
	Time          time.Time
	Action        string
	InitiatorID   string
	AffectedIDs   []string
	Decision      string // "allowed"|"blocked"|"throttled"
	TrustChange   float64
	EthicsChange  float64
	LegitimacyRef string
}

// TrustStore is the pluggable persistence for trust/ethics scores.
type TrustStore interface {
	GetTrust(ctx context.Context, actorID string) (float64, bool, error)
	SetTrust(ctx context.Context, actorID string, v float64) error
	GetEthics(ctx context.Context, actorID string) (float64, bool, error)
	SetEthics(ctx context.Context, actorID string, v float64) error
}

// Simulator receives feedback events (e.g., Simula Terra).
type Simulator interface {
	OnFeedback(ctx context.Context, evt Event) error
}

// Event is emitted when any score changes.
type Event struct {
	ActorID       string
	Action        string
	Decision      string
	TrustAfter    float64
	EthicsAfter   float64
	ReceiptID     string
	LegitimacyRef string
	Time          time.Time
}

// Loop applies moral feedback from receipts.
type Loop struct {
	store     TrustStore
	sim       Simulator
	mu        sync.Mutex
	softFloor float64
	softCeil  float64
}

// NewLoop creates a feedback loop with soft clamps on scores.
func NewLoop(store TrustStore, sim Simulator) *Loop {
	return &Loop{
		store:     store,
		sim:       sim,
		softFloor: 0.0,
		softCeil:  1.0,
	}
}

// Apply implements the consent.FeedbackSink interface semantics.
func (l *Loop) Apply(ctx context.Context, rcpt Receipt) error {
	// Initiator always gets updated.
	if err := l.bump(ctx, rcpt.InitiatorID, rcpt); err != nil {
		return err
	}
	// Affected parties may also shift ethics perception (small echo).
	for _, aid := range rcpt.AffectedIDs {
		echo := rcpt
		// Halve the magnitude for affected parties to avoid runaway effects.
		echo.TrustChange *= 0.5
		echo.EthicsChange *= 0.5
		if err := l.bump(ctx, aid, echo); err != nil {
			// non-fatal in baseline
			continue
		}
	}
	return nil
}

func (l *Loop) bump(ctx context.Context, actorID string, rcpt Receipt) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	t0, ok, err := l.store.GetTrust(ctx, actorID)
	if err != nil {
		return err
	}
	if !ok {
		t0 = 0.5 // neutral default
	}

	e0, ok, err := l.store.GetEthics(ctx, actorID)
	if err != nil {
		return err
	}
	if !ok {
		e0 = 0.5
	}

	t1 := clamp(t0+rcpt.TrustChange, l.softFloor, l.softCeil)
	e1 := clamp(e0+rcpt.EthicsChange, l.softFloor, l.softCeil)

	if err := l.store.SetTrust(ctx, actorID, t1); err != nil {
		return err
	}
	if err := l.store.SetEthics(ctx, actorID, e1); err != nil {
		return err
	}

	if l.sim != nil {
		_ = l.sim.OnFeedback(ctx, Event{
			ActorID:       actorID,
			Action:        rcpt.Action,
			Decision:      rcpt.Decision,
			TrustAfter:    t1,
			EthicsAfter:   e1,
			ReceiptID:     rcpt.ID,
			LegitimacyRef: rcpt.LegitimacyRef,
			Time:          rcpt.Time,
		})
	}
	return nil
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

// --------- In-memory store (baseline) ---------

type MemStore struct {
	mu     sync.RWMutex
	trust  map[string]float64
	ethics map[string]float64
}

func NewMemStore() *MemStore {
	return &MemStore{
		trust:  make(map[string]float64),
		ethics: make(map[string]float64),
	}
}

func (m *MemStore) GetTrust(ctx context.Context, actorID string) (float64, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.trust[actorID]
	return v, ok, nil
}

func (m *MemStore) SetTrust(ctx context.Context, actorID string, v float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.trust[actorID] = v
	return nil
}

func (m *MemStore) GetEthics(ctx context.Context, actorID string) (float64, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.ethics[actorID]
	return v, ok, nil
}

func (m *MemStore) SetEthics(ctx context.Context, actorID string, v float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ethics[actorID] = v
	return nil
}
