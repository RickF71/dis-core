package identity

import (
	"context"
	"fmt"

	"dis-core/internal/core/authority"
	"dis-core/internal/core/domain"
	"dis-core/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// -----------------------------------------------------------------------------
// ResolveInviteSubject
// -----------------------------------------------------------------------------

func ResolveInviteSubject(ctx context.Context, tx pgx.Tx, token string) (string, error) {
	var subject string
	if err := tx.QueryRow(ctx,
		`SELECT subject FROM handshakes WHERE token = $1`,
		token,
	).Scan(&subject); err != nil {
		return "", fmt.Errorf("resolve invite subject: %w", err)
	}
	return subject, nil
}

// -----------------------------------------------------------------------------
// KnowThyselfAtomic
// -----------------------------------------------------------------------------
//
// New behavior:
//   1. Resolve invite subject
//   2. Create corporeal identity w/ presentation name
//   3. Create corporeal domain w/ same presentation name
//   4. Instantiate prime seat for (actor, domain)
//   5. Record receipt: ci.call.login.v1
//   6. Return both IDs
//

type KnowThyselfResult struct {
	ActorID   uuid.UUID
	DomainID  uuid.UUID
	ReceiptID uuid.UUID
}

func KnowThyselfAtomic(
	ctx context.Context,
	tx pgx.Tx,
	inviteToken string,
	presentationName string,
) (*KnowThyselfResult, error) {

	if presentationName == "" {
		return nil, fmt.Errorf("presentation name is required")
	}

	// 1) Resolve invite subject
	subject, err := ResolveInviteSubject(ctx, tx, inviteToken)
	if err != nil {
		return nil, fmt.Errorf("know_thyself: resolve invite: %w", err)
	}

	// 2) Create corporeal identity
	actorID := uuid.New()
	_, err = tx.Exec(ctx,
		`INSERT INTO identities (id, subject, presentation_name, identity_type)
		 VALUES ($1, $2, $3, 'corporeal')`,
		actorID, subject, presentationName,
	)
	if err != nil {
		return nil, fmt.Errorf("know_thyself: create identity: %w", err)
	}

	// 3) Create corporeal domain using canonical helper
	//
	// We reuse your existing domain helper but pass presentationName
	// because it now *must* become the user domain name.
	//
	domainID, err := domain.CreateCorporealDomainNamed(
		ctx,
		tx,
		actorID,
		presentationName,
	)
	if err != nil {
		return nil, fmt.Errorf("know_thyself: create domain: %w", err)
	}

	// 4) Instantiate prime seat using canonical domain_seats helper
	seatID := uuid.New()
	if err := db.InsertSeatTx(ctx, tx, seatID.String(), domainID.String(), actorID.String(), "prime"); err != nil {
		return nil, fmt.Errorf("know_thyself: create pseat: %w", err)
	}

	// 5) Record genesis login receipt
	payload := map[string]any{
		"domain":         domainID.String(),
		"actor":          actorID.String(),
		"presentation":   presentationName,
		"issued_by":      "domain.null", // your existing rule
		"invite_subject": subject,
		"invite_token":   inviteToken,
	}

	receiptIDStr, err := authority.RecordAuthorityReceiptTx(
		ctx,
		tx,
		domainID.String(),
		"ci.call.login.v1",
		payload,
	)
	if err != nil {
		return nil, fmt.Errorf("know_thyself: record receipt: %w", err)
	}
	receiptID, _ := uuid.Parse(receiptIDStr)

	// 6) Return both IDs
	return &KnowThyselfResult{
		ActorID:   actorID,
		DomainID:  domainID,
		ReceiptID: receiptID,
	}, nil
}
