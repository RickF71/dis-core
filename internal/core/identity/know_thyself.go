package identity

import (
	"context"
	"fmt"

	"dis-core/internal/core/authority"
	"dis-core/internal/core/domain"
	"dis-core/internal/db"
	"dis-core/internal/util"

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

	// 1) Resolve invite subject and lock the handshake row so concurrent
	// accept attempts are serialized. We also read the handshake status so
	// we can support a minimal genesis/bootstrap handshake (status='genesis')
	// that may have an empty subject and use the provided presentation name
	// as the resolved subject.
	var subject string
	var status string
	var err error
	if err = tx.QueryRow(ctx, `SELECT subject, COALESCE(status,'') FROM handshakes WHERE token = $1 FOR UPDATE`, inviteToken).Scan(&subject, &status); err != nil {
		return nil, fmt.Errorf("know_thyself: resolve invite: %w", err)
	}

	// If this is a genesis/bootstrap handshake (status='genesis') and the
	// stored subject is empty, we treat this as the minimal bootstrap path:
	// - subject := presentationName
	// - create identity
	// - ensure domain 'domain.null' exists
	// - instantiate/assign prime seat for that domain to this identity
	if subject == "" && status == "genesis" {
		// 2a) Use presentationName as subject for bootstrap
		subject = presentationName

		// create identity (corporeal for consistency with other flows)
		actorID := uuid.New()
		_, err = tx.Exec(ctx,
			`INSERT INTO identities (id, subject, presentation_name, identity_type)
			 VALUES ($1, $2, $3, 'corporeal')`,
			actorID, subject, presentationName,
		)
		if err != nil {
			return nil, fmt.Errorf("know_thyself: create identity (bootstrap): %w", err)
		}

		// ensure domain.null exists
		nullDomainID, err := domain.EnsureNullDomainExistsTx(ctx, tx)
		if err != nil {
			return nil, fmt.Errorf("know_thyself: ensure null domain: %w", err)
		}

		// instantiate prime seat assigned to this identity
		seatID := uuid.New()
		if err := db.InsertSeatTx(ctx, tx, seatID.String(), nullDomainID.String(), actorID.String(), "prime"); err != nil {
			return nil, fmt.Errorf("know_thyself: create null pseat: %w", err)
		}

		// 5a) Record genesis login receipt
		payload := map[string]any{
			"domain":       nullDomainID.String(),
			"actor":        actorID.String(),
			"presentation": presentationName,
			"issued_by":    "domain.null",
			"invite_token": inviteToken,
			"roles":        []string{},
		}

		receiptIDStr, err := authority.RecordAuthorityReceiptTx(
			ctx,
			tx,
			nullDomainID.String(),
			"ci.call.login.v1",
			payload,
		)
		if err != nil {
			return nil, fmt.Errorf("know_thyself: record receipt (bootstrap): %w", err)
		}
		receiptID, _ := uuid.Parse(receiptIDStr)

		// consume handshake
		if _, err := tx.Exec(ctx, `DELETE FROM handshakes WHERE token = $1`, inviteToken); err != nil {
			return nil, fmt.Errorf("know_thyself: consume handshake (bootstrap): %w", err)
		}

		// Defensive validation
		if err := util.ValidateDomainUID(nullDomainID.String()); err != nil {
			return nil, fmt.Errorf("know_thyself: created domain id invalid: %w", err)
		}

		return &KnowThyselfResult{
			ActorID:   actorID,
			DomainID:  nullDomainID,
			ReceiptID: receiptID,
		}, nil
	}

	// 2) Create corporeal identity (normal flow)
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
	// Include roles explicitly (empty list at genesis) and ensure the
	// domain/actor fields use UID strings only.
	payload := map[string]any{
		"domain":         domainID.String(),
		"actor":          actorID.String(),
		"presentation":   presentationName,
		"issued_by":      "domain.null",
		"invite_subject": subject,
		"invite_token":   inviteToken,
		"roles":          []string{},
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

	// 6) Consume the handshake so the token cannot be reused
	if _, err := tx.Exec(ctx, `DELETE FROM handshakes WHERE token = $1`, inviteToken); err != nil {
		return nil, fmt.Errorf("know_thyself: consume handshake: %w", err)
	}

	// Defensive validation: ensure the created domain ID is a valid UUID
	// (it will be because we generated it, but this keeps invariant checks
	// explicit and mirrors the UID-only enforcement for inputs).
	if err := util.ValidateDomainUID(domainID.String()); err != nil {
		return nil, fmt.Errorf("know_thyself: created domain id invalid: %w", err)
	}

	return &KnowThyselfResult{
		ActorID:   actorID,
		DomainID:  domainID,
		ReceiptID: receiptID,
	}, nil
}
