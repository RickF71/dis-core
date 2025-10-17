package events_terra

import (
	"context"
	"errors"
	"time"
)

type Signature string

type GiveRootEvent struct {
	ID           string
	GiverID      string
	ApplicantID  string
	MethodKind   string // "mutual_recognition" | "adapted_recognition"
	MethodNote   string
	Timestamp    time.Time
	GiverSig     Signature
	ApplicantSig Signature
}

func (e *GiveRootEvent) ValidateBasic() error {
	if e.GiverID == "" || e.ApplicantID == "" {
		return errors.New("missing ids")
	}
	if e.GiverSig == "" || e.ApplicantSig == "" {
		return errors.New("missing signatures")
	}
	if e.MethodKind == "" {
		return errors.New("method required")
	}
	return nil
}

type TerraLedger interface {
	IsRooted(ctx context.Context, personID string) (bool, error)
	CommitGiveRoot(ctx context.Context, e *GiveRootEvent) (receiptHash string, err error)
}

func CommitGiveRoot(ctx context.Context, ledger TerraLedger, e *GiveRootEvent) (string, error) {
	if err := e.ValidateBasic(); err != nil {
		return "", err
	}
	// giver must already be rooted
	rooted, err := ledger.IsRooted(ctx, e.GiverID)
	if err != nil {
		return "", err
	}
	if !rooted {
		return "", errors.New("giver not rooted")
	}
	return ledger.CommitGiveRoot(ctx, e) // atomic inside the ledger impl
}
