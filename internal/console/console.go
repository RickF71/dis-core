package console

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"dis-core/internal/db"
	"dis-core/internal/receipts"
)

// Console represents an Authority Console instance bound to a domain & core.
type Console struct {
	ID          string
	BoundDomain string
	BoundCore   string
	SeatHolders []string
	Actions     []ConsoleAction
}

// ConsoleAction represents a single policy or operational act recorded by the console.
type ConsoleAction struct {
	ID        string
	Type      string
	PolicyRef string
	CreatedAt string
	Initiator string
	Status    string
	Reason    string
	Receipt   *receipts.Receipt
}

// NewConsole initializes a new Authority Console.
func NewConsole(boundDomain, boundCore string, seats []string) *Console {
	b := make([]byte, 4)
	rand.Read(b)
	consoleID := fmt.Sprintf("ac-%s", hex.EncodeToString(b))

	log.Printf("⚙️  Created Authority Console: %s (domain=%s core=%s)", consoleID, boundDomain, boundCore)

	return &Console{
		ID:          consoleID,
		BoundDomain: boundDomain,
		BoundCore:   boundCore,
		SeatHolders: seats,
	}
}

// LogAction creates and records a new action and its signed receipt.
func (c *Console) LogAction(actionType, policyRef, initiator string) (*ConsoleAction, error) {
	// Validate initiator
	valid := false
	for _, s := range c.SeatHolders {
		if s == initiator {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("unauthorized initiator: %s", initiator)
	}

	createdAt := db.NowRFC3339Nano()
	actionID := generateActionID()

	// Generate the receipt via receipts.NewReceipt
	r := receipts.NewReceipt(c.BoundDomain, actionType, c.BoundCore, c.ID, initiator)

	// Save the receipt to disk under the current version’s receipt folder
	saveDir := "versions/v0.6/receipts/generated"
	if err := r.Save(saveDir); err != nil {
		return nil, fmt.Errorf("failed to save receipt: %v", err)
	}

	act := ConsoleAction{
		ID:        actionID,
		Type:      actionType,
		PolicyRef: policyRef,
		CreatedAt: createdAt,
		Initiator: initiator,
		Status:    "executed",
		Receipt:   r,
	}

	c.Actions = append(c.Actions, act)
	log.Printf("🧾 Action logged: %s (%s)\n", actionType, actionID)
	return &act, nil
}

func generateActionID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return fmt.Sprintf("act-%s", hex.EncodeToString(b))
}
