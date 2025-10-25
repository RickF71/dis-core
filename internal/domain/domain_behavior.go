package domain

import (
	"log"
	"sync"
	"time"

	"dis-core/internal/events"
	"dis-core/internal/ledger"
	"dis-core/internal/rules"
)

// DomainBrain represents a simple cognitive loop for a domain.
// It continuously polls for events, interprets them using its BehaviorRules,
// and emits reflexive receipts when thresholds are crossed.
type DomainBrain struct {
	DomainID   string
	Trust      float64
	Ethics     float64
	ReflexRate time.Duration
	Ruleset    *rules.BehaviorSet
	stop       chan struct{}
	wg         sync.WaitGroup
}

// NewDomainBrain initializes a new domain behavior loop.
func NewDomainBrain(id string, ruleset *rules.BehaviorSet) *DomainBrain {
	return &DomainBrain{
		DomainID:   id,
		Trust:      1.0,
		Ethics:     1.0,
		ReflexRate: 3 * time.Second,
		Ruleset:    ruleset,
		stop:       make(chan struct{}),
	}
}

// Start begins the loop as a goroutine.
func (b *DomainBrain) Start() {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		ticker := time.NewTicker(b.ReflexRate)
		for {
			select {
			case <-ticker.C:
				b.processCycle()
			case <-b.stop:
				ticker.Stop()
				log.Printf("🧠 DomainBrain stopped for %s", b.DomainID)
				return
			}
		}
	}()
	log.Printf("🧠 DomainBrain started for %s", b.DomainID)
}

// Stop cleanly halts the loop.
func (b *DomainBrain) Stop() {
	close(b.stop)
	b.wg.Wait()
}

func (b *DomainBrain) processCycle() {
	evts := events.FetchRecent(b.DomainID)
	for _, e := range evts {
		action := b.Ruleset.Decide(e)
		if action.Receipt {
			ledger.EmitReflexiveReceipt(b.DomainID, e, action)
		}
		b.updateState(action)
	}
}

func (b *DomainBrain) updateState(a rules.Action) {
	b.Trust += a.TrustDelta
	b.Ethics += a.EthicsDelta

	// Apply soft bounds
	if b.Trust > 1.0 {
		b.Trust = 1.0
	}
	if b.Trust < -1.0 {
		b.Trust = -1.0
	}
	if b.Ethics > 1.0 {
		b.Ethics = 1.0
	}
	if b.Ethics < -1.0 {
		b.Ethics = -1.0
	}

	log.Printf("⚖️ %s Trust: %.3f Ethics: %.3f", b.DomainID, b.Trust, b.Ethics)
}
