package seats

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles database operations for seats
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new seats repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetPrimeSeat retrieves the Prime Seat for a domain
func (r *Repository) GetPrimeSeat(ctx context.Context, domainID uuid.UUID) (*Seat, error) {
	var seat Seat
	query := `
		SELECT id, domain_id, parent_seat_id, seat_type, member_id,
		       appointed_by, appointment_receipt, rego_ref, rego_text,
		       policy_version, scope, status, created_at
		FROM domain_seats
		WHERE domain_id = $1 AND seat_type = 'prime'
		LIMIT 1
	`
	err := r.db.QueryRow(ctx, query, domainID).Scan(
		&seat.ID, &seat.DomainID, &seat.ParentSeatID, &seat.SeatType,
		&seat.MemberID, &seat.AppointedBy, &seat.AppointmentReceipt,
		&seat.RegoRef, &seat.RegoText, &seat.PolicyVersion, &seat.Scope,
		&seat.Status, &seat.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get prime seat: %w", err)
	}
	return &seat, nil
}

// GOV-7: CreatePrimeSeat creates a new Prime Seat for a corporeal domain
// Enforces single-occupancy and non-transferability
func (r *Repository) CreatePrimeSeat(ctx context.Context, domainID uuid.UUID, memberID, regoRef string) (*Seat, error) {
	seat := &Seat{
		ID:       uuid.New(),
		DomainID: domainID,
		SeatType: "prime",
		MemberID: &memberID,
		Status:   "active",
	}

	if regoRef != "" {
		seat.RegoRef = &regoRef
	}

	query := `
		INSERT INTO domain_seats (
			id, domain_id, seat_type, member_id, rego_ref, status
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	err := r.db.QueryRow(ctx, query,
		seat.ID, seat.DomainID, seat.SeatType, seat.MemberID, seat.RegoRef, seat.Status,
	).Scan(&seat.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("create prime seat: %w", err)
	}

	return seat, nil
}

// GetActiveSeats retrieves all active seats for a domain
func (r *Repository) GetActiveSeats(ctx context.Context, domainID uuid.UUID) ([]Seat, error) {
	query := `
		SELECT id, domain_id, parent_seat_id, seat_type, member_id,
		       appointed_by, appointment_receipt, rego_ref, rego_text,
		       policy_version, scope, status, created_at
		FROM domain_seats
		WHERE domain_id = $1 AND status = 'active'
		ORDER BY seat_type DESC, created_at ASC
	`
	rows, err := r.db.Query(ctx, query, domainID)
	if err != nil {
		return nil, fmt.Errorf("get active seats: %w", err)
	}
	defer rows.Close()

	var seats []Seat
	for rows.Next() {
		var seat Seat
		err := rows.Scan(
			&seat.ID, &seat.DomainID, &seat.ParentSeatID, &seat.SeatType,
			&seat.MemberID, &seat.AppointedBy, &seat.AppointmentReceipt,
			&seat.RegoRef, &seat.RegoText, &seat.PolicyVersion, &seat.Scope,
			&seat.Status, &seat.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan seat: %w", err)
		}
		seats = append(seats, seat)
	}

	return seats, nil
}

// AppointMemberSeat creates a new member seat appointment
func (r *Repository) AppointMemberSeat(ctx context.Context, domainID uuid.UUID, pseatID uuid.UUID, memberID string,
	scope, regoRef, policyVersion, appointmentReceipt string) (*Seat, error) {

	seat := &Seat{
		ID:                 uuid.New(),
		DomainID:           domainID,
		ParentSeatID:       &pseatID,
		SeatType:           "member",
		MemberID:           &memberID,
		AppointedBy:        &pseatID,
		AppointmentReceipt: &appointmentReceipt,
		RegoRef:            &regoRef,
		Scope:              &scope,
		PolicyVersion:      &policyVersion,
		Status:             "active",
	}

	query := `
		INSERT INTO domain_seats (
			id, domain_id, parent_seat_id, seat_type, member_id, appointed_by,
			appointment_receipt, rego_ref, scope, policy_version, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at
	`

	err := r.db.QueryRow(ctx, query,
		seat.ID, seat.DomainID, seat.ParentSeatID, seat.SeatType, seat.MemberID,
		seat.AppointedBy, seat.AppointmentReceipt, seat.RegoRef, seat.Scope,
		seat.PolicyVersion, seat.Status,
	).Scan(&seat.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("appoint member seat: %w", err)
	}

	return seat, nil
}

// FreezeSeat sets a seat's status to frozen
func (r *Repository) FreezeSeat(ctx context.Context, seatID uuid.UUID) error {
	query := `UPDATE domain_seats SET status = 'frozen' WHERE id = $1`
	_, err := r.db.Exec(ctx, query, seatID)
	if err != nil {
		return fmt.Errorf("freeze seat: %w", err)
	}
	return nil
}

// UnfreezeSeat sets a seat's status back to active
func (r *Repository) UnfreezeSeat(ctx context.Context, seatID uuid.UUID) error {
	query := `UPDATE domain_seats SET status = 'active' WHERE id = $1`
	_, err := r.db.Exec(ctx, query, seatID)
	if err != nil {
		return fmt.Errorf("unfreeze seat: %w", err)
	}
	return nil
}

// UpdateSeatRego updates the per-seat REGO policy
func (r *Repository) UpdateSeatRego(ctx context.Context, seatID uuid.UUID, regoText, policyVersion string) error {
	query := `
		UPDATE domain_seats
		SET rego_text = $2, policy_version = $3
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, seatID, regoText, policyVersion)
	if err != nil {
		return fmt.Errorf("update seat rego: %w", err)
	}
	return nil
}

// GetSeat retrieves a single seat by ID
func (r *Repository) GetSeat(ctx context.Context, seatID uuid.UUID) (*Seat, error) {
	var seat Seat
	query := `
		SELECT id, domain_id, parent_seat_id, seat_type, member_id,
		       appointed_by, appointment_receipt, rego_ref, rego_text,
		       policy_version, scope, status, created_at
		FROM domain_seats
		WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, seatID).Scan(
		&seat.ID, &seat.DomainID, &seat.ParentSeatID, &seat.SeatType,
		&seat.MemberID, &seat.AppointedBy, &seat.AppointmentReceipt,
		&seat.RegoRef, &seat.RegoText, &seat.PolicyVersion, &seat.Scope,
		&seat.Status, &seat.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get seat: %w", err)
	}
	return &seat, nil
}

// EnsurePrimeSeat creates a Prime Seat for a domain if it doesn't exist
func (r *Repository) EnsurePrimeSeat(ctx context.Context, domainID uuid.UUID) (*Seat, error) {
	// Check if Prime Seat exists
	existing, err := r.GetPrimeSeat(ctx, domainID)
	if err == nil {
		return existing, nil
	}

	// Create Prime Seat
	scopeVal := "domain"
	seat := &Seat{
		ID:       uuid.New(),
		DomainID: domainID,
		SeatType: "prime",
		Status:   "active",
		Scope:    &scopeVal,
	}

	query := `
		INSERT INTO domain_seats (id, domain_id, seat_type, status, scope)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`

	err = r.db.QueryRow(ctx, query, seat.ID, seat.DomainID, seat.SeatType, seat.Status, seat.Scope).
		Scan(&seat.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("create prime seat: %w", err)
	}

	return seat, nil
}

// VerifySeatOwnership verifies that a seat belongs to a user identified by externalUID
// Checks if the seat's member_id matches patterns: human.{uid}, actor.{uid}, or direct match
func (r *Repository) VerifySeatOwnership(ctx context.Context, seatID uuid.UUID, externalUID string) error {
	if externalUID == "" {
		return fmt.Errorf("external UID is required")
	}

	query := `
		SELECT 1
		FROM domain_seats
		WHERE id = $1
		  AND (
		    member_id LIKE 'human.' || $2 || '%'
		    OR member_id LIKE 'actor.' || $2 || '%'
		    OR member_id = $2
		  )
		  AND status = 'active'
		LIMIT 1
	`

	var exists int
	err := r.db.QueryRow(ctx, query, seatID, externalUID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("seat not found or not owned by user: %w", err)
	}

	return nil
}
