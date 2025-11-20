package authority

import "time"

// ReceiptSeatInstantiateV1 is a MinSet-5 style receipt produced when a seat
// is instantiated. Fields are intentionally minimal and will be extended
// as MX develops.
type ReceiptSeatInstantiateV1 struct {
	ReceiptID     string    `json:"receipt_id"`
	SeatID        string    `json:"seat_id"`
	DomainID      string    `json:"domain_id"`
	IdentityID    string    `json:"identity_id"`
	Timestamp     time.Time `json:"timestamp"`
	PolicyRef     string    `json:"policy_ref,omitempty"`
	ProvenanceRef string    `json:"provenance_ref,omitempty"`
}

// ReceiptSeatLineageQueryV1 and ReceiptSeatStatusV1 are placeholders for
// lineage and status receipts; they will be filled in as needed.
type ReceiptSeatLineageQueryV1 struct {
	ReceiptID string    `json:"receipt_id"`
	SeatID    string    `json:"seat_id"`
	Timestamp time.Time `json:"timestamp"`
}

type ReceiptSeatStatusV1 struct {
	ReceiptID string    `json:"receipt_id"`
	SeatID    string    `json:"seat_id"`
	Timestamp time.Time `json:"timestamp"`
	Active    bool      `json:"active"`
}
