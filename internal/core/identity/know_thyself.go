package identity

import (
	"context"
	"errors"
	"fmt"

	"dis-core/internal/core/authority"
	"dis-core/internal/core/bedrock"
	"dis-core/internal/core/domain"
	"dis-core/internal/db"
	"dis-core/internal/receipts"
	"dis-core/internal/util"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrUnknownInvite indicates the provided invite token does not exist.
var ErrUnknownInvite = errors.New("unknown invite token")

// expanded back to the full spec after build/tests are validated.

func ResolveInviteSubject(ctx context.Context, tx pgx.Tx, token string) (string, error) {
	var subject string
	err := tx.QueryRow(ctx, `SELECT subject FROM handshakes WHERE token = $1`, token).Scan(&subject)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", ErrUnknownInvite
		}
		return "", fmt.Errorf("resolve invite subject: %w", err)
	}
	return subject, nil
}

type KnowThyselfResult struct {
	IdentityID uuid.UUID
	DomainID   uuid.UUID
	ActorID    uuid.UUID
	PseatID    uuid.UUID
	ReceiptID  uuid.UUID
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

	if booted, err := bedrock.IsSystemBootstrapped(ctx, tx); err != nil {
		return nil, fmt.Errorf("know_thyself: check bootstrapped: %w", err)
	} else if !booted {
		if err := bedrock.BedrockBootstrapAtomic(ctx, tx); err != nil {
			return nil, fmt.Errorf("know_thyself: bedrock bootstrap: %w", err)
		}
	}

	// Resolve invite subject and load handshake status in a single FOR UPDATE
	// query to lock the handshake row and avoid visibility/isolation issues.
	var origSubject string
	var status string
	if err := tx.QueryRow(ctx, `SELECT subject, COALESCE(status,'') FROM handshakes WHERE token = $1 FOR UPDATE`, inviteToken).Scan(&origSubject, &status); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUnknownInvite
		}
		return nil, fmt.Errorf("know_thyself: resolve invite: %w", err)
	}

	subject := origSubject
	if subject == "" {
		if status == "genesis" {
			subject = presentationName
		} else {
			return nil, fmt.Errorf("invite subject invalid")
		}
	}

	if origSubject == "" && status == "genesis" {
		actorID := uuid.New()
		if _, err := tx.Exec(ctx,
			`INSERT INTO identities (id, subject, presentation_name, identity_type)
			 VALUES ($1, $2, $3, 'corporeal')`,
			actorID, subject, presentationName,
		); err != nil {
			return nil, fmt.Errorf("know_thyself: create identity (bootstrap): %w", err)
		}

		nullDomainID, err := domain.EnsureNullDomainExistsTx(ctx, tx)
		if err != nil {
			return nil, fmt.Errorf("know_thyself: ensure null domain: %w", err)
		}

		// Ensure or create the null prime seat and bind it to this new identity.
		seatID, err := domain.EnsureNullPrimeSeatExistsTx(ctx, tx, nullDomainID, actorID.String())
		if err != nil {
			return nil, fmt.Errorf("know_thyself: ensure null prime seat: %w", err)
		}

		// Mark seat roles required by AT-0
		if err := db.AddRoleTx(ctx, tx, seatID.String(), "sovereign"); err != nil {
			return nil, fmt.Errorf("know_thyself: add role sovereign: %w", err)
		}
		if err := db.AddRoleTx(ctx, tx, seatID.String(), "immutable"); err != nil {
			return nil, fmt.Errorf("know_thyself: add role immutable: %w", err)
		}
		if err := db.AddRoleTx(ctx, tx, seatID.String(), "nonrevocable"); err != nil {
			return nil, fmt.Errorf("know_thyself: add role nonrevocable: %w", err)
		}

		payload := map[string]any{
			"domain_id":   nullDomainID.String(),
			"seat_id":     seatID.String(),
			"identity_id": actorID.String(),
			"created_by":  "know_thyself",
			"policy_ref":  "",
		}

		receiptIDStr, err := authority.RecordAuthorityReceiptTx(ctx, tx, nullDomainID.String(), "pseat.instantiate.null.v1", payload)
		if err != nil {
			return nil, fmt.Errorf("know_thyself: record pseat instantiation receipt: %w", err)
		}
		receiptID, _ := uuid.Parse(receiptIDStr)

		if _, err := tx.Exec(ctx, `DELETE FROM handshakes WHERE token = $1`, inviteToken); err != nil {
			return nil, fmt.Errorf("know_thyself: consume handshake (bootstrap): %w", err)
		}

		if err := util.ValidateDomainUID(nullDomainID.String()); err != nil {
			return nil, fmt.Errorf("know_thyself: created domain id invalid: %w", err)
		}

		// Bind bootstrap actor if not already set. This ensures the first
		// corporeal identity created during bootstrap becomes the canonical
		// bootstrap actor for policy rules that gate root operations.
		if err := db.SetBootstrapActorIDTx(ctx, tx, actorID.String()); err != nil {
			return nil, fmt.Errorf("know_thyself: set bootstrap actor id: %w", err)
		}

		return &KnowThyselfResult{
			IdentityID: actorID,
			DomainID:   nullDomainID,
			ActorID:    actorID,
			PseatID:    seatID,
			ReceiptID:  receiptID,
		}, nil
	}

	actorID := uuid.New()
	if _, err := tx.Exec(ctx,
		`INSERT INTO identities (id, subject, presentation_name, identity_type)
		 VALUES ($1, $2, $3, 'corporeal')`,
		actorID, subject, presentationName,
	); err != nil {
		return nil, fmt.Errorf("know_thyself: create identity: %w", err)
	}

	domainID, err := domain.CreateCorporealDomainNamed(ctx, tx, actorID, presentationName)
	if err != nil {
		return nil, fmt.Errorf("know_thyself: create domain: %w", err)
	}

	seatID := uuid.New()
	if err := db.InsertSeatTx(ctx, tx, seatID.String(), domainID.String(), actorID.String(), "prime"); err != nil {
		return nil, fmt.Errorf("know_thyself: create pseat: %w", err)
	}

	// Root-seat assignment: if no root sovereign exists in the null domain,
	// seat this new corporeal pactor on the null domain's prime seat and
	// emit a root.sovereign.established.v1 receipt for provenance.
	// Resolve null domain ID (ensure exists) and check for existing root.
	nullDomainID, err := domain.EnsureNullDomainExistsTx(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("know_thyself: ensure null domain for root-seat: %w", err)
	}
	hasRoot, err := domain.HasRootSovereignTx(ctx, tx, nullDomainID)
	if err != nil {
		return nil, fmt.Errorf("know_thyself: check root sovereign: %w", err)
	}
	if !hasRoot {
		// Create a new prime seat in the null domain and assign it to this pactor.
		newRootSeatID := uuid.New()
		if err := db.InsertSeatTx(ctx, tx, newRootSeatID.String(), nullDomainID.String(), actorID.String(), "prime"); err != nil {
			return nil, fmt.Errorf("root seat assignment: %w", err)
		}
		// Attach minimal sovereign role to the root seat.
		if err := db.AddRoleTx(ctx, tx, newRootSeatID.String(), "sovereign"); err != nil {
			return nil, fmt.Errorf("root seat role assignment: %w", err)
		}

		// Emit a provenance receipt indicating the root sovereign was established.
		if err := receipts.EmitRootSovereignEstablished(ctx, tx, receipts.RootSovereignEstablished{
			RootDomainID:      nullDomainID.String(),
			CorporealDomainID: domainID.String(),
			PactorID:          actorID.String(),
			IssuedBy:          "0D-code",
			CreatedAt:         time.Now().UTC(),
		}); err != nil {
			return nil, fmt.Errorf("emit root sovereign receipt: %w", err)
		}
	}

	payload := map[string]any{
		"domain":         domainID.String(),
		"actor":          actorID.String(),
		"presentation":   presentationName,
		"issued_by":      "domain.null",
		"invite_subject": subject,
		"invite_token":   inviteToken,
		"roles":          []string{},
	}

	receiptIDStr, err := authority.RecordAuthorityReceiptTx(ctx, tx, domainID.String(), "ci.call.login.v1", payload)
	if err != nil {
		return nil, fmt.Errorf("know_thyself: record receipt: %w", err)
	}
	receiptID, _ := uuid.Parse(receiptIDStr)

	if _, err := tx.Exec(ctx, `DELETE FROM handshakes WHERE token = $1`, inviteToken); err != nil {
		return nil, fmt.Errorf("know_thyself: consume handshake: %w", err)
	}

	if err := util.ValidateDomainUID(domainID.String()); err != nil {
		return nil, fmt.Errorf("know_thyself: created domain id invalid: %w", err)
	}

	return &KnowThyselfResult{
		IdentityID: actorID,
		DomainID:   domainID,
		ActorID:    actorID,
		PseatID:    uuid.Nil,
		ReceiptID:  receiptID,
	}, nil
}

//
