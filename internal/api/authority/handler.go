package authorityapi

import (
	"encoding/json"
	"net/http"

	"dis-core/internal/authority"
	"dis-core/internal/identity"

	"github.com/go-chi/chi/v5"
)

// TriadHandler handles identity triad API requests
type TriadHandler struct {
	triadRepo      *identity.TriadRepository
	mutationEngine *authority.SeatMutationEngine // GOV-3: seat mutation engine
}

// NewTriadHandler creates a new triad API handler
func NewTriadHandler(triadRepo *identity.TriadRepository) *TriadHandler {
	return &TriadHandler{
		triadRepo:      triadRepo,
		mutationEngine: nil, // Set via SetMutationEngine after construction
	}
}

// SetMutationEngine sets the mutation engine (GOV-3)
func (h *TriadHandler) SetMutationEngine(engine *authority.SeatMutationEngine) {
	h.mutationEngine = engine
}

// GetTriadByIdentity returns the identity triad for a specific identity
// GET /api/authority/triad/:identityId
func (h *TriadHandler) GetTriadByIdentity(w http.ResponseWriter, r *http.Request) {
	identityID := chi.URLParam(r, "identityId")
	if identityID == "" {
		http.Error(w, "missing identityId", http.StatusBadRequest)
		return
	}

	triad, err := h.triadRepo.GetIdentityTriad(r.Context(), identityID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Convert to DTO
	dto := &IdentityTriadDTO{
		IdentityID: identityID,
		Complete:   triad.IsComplete(),
		Seats:      make([]IdentitySeatDTO, 0, 3),
	}

	if triad.Terra != nil {
		dto.Seats = append(dto.Seats, IdentitySeatDTO{
			SeatID:    triad.Terra.ID.String(),
			Layer:     "terra",
			State:     triad.Terra.State,
			UpdatedAt: formatTime(triad.Terra.CreatedAt),
		})
	}

	if triad.Numen != nil {
		dto.Seats = append(dto.Seats, IdentitySeatDTO{
			SeatID:    triad.Numen.ID.String(),
			Layer:     "numen",
			State:     triad.Numen.State,
			UpdatedAt: formatTime(triad.Numen.CreatedAt),
		})
	}

	if triad.Lima != nil {
		dto.Seats = append(dto.Seats, IdentitySeatDTO{
			SeatID:    triad.Lima.ID.String(),
			Layer:     "lima",
			State:     triad.Lima.State,
			UpdatedAt: formatTime(triad.Lima.CreatedAt),
		})
	}

	writeJSON(w, dto)
}

// PreviewFlow returns sample authority flow scenarios
// GET /api/authority/flow/preview
func (h *TriadHandler) PreviewFlow(w http.ResponseWriter, r *http.Request) {
	preview := &FlowPreview{
		Timestamp: GetTimestamp(),
		Scenarios: []FlowScenario{
			{
				Name:        "Upward Reporting (ASSIGNED)",
				Description: "Identity with ASSIGNED seat can report upward",
				Direction:   "upward",
				SeatState:   "ASSIGNED",
				Expected:    "allow",
			},
			{
				Name:        "Upward Reporting (OCCUPIED)",
				Description: "Identity with OCCUPIED seat can report upward",
				Direction:   "upward",
				SeatState:   "OCCUPIED",
				Expected:    "allow",
			},
			{
				Name:        "Downward Governance (OCCUPIED)",
				Description: "Identity with OCCUPIED seat can govern downward within domain",
				Direction:   "downward",
				SeatState:   "OCCUPIED",
				Expected:    "allow",
			},
			{
				Name:        "Downward Governance (ASSIGNED)",
				Description: "Identity with ASSIGNED seat cannot govern downward",
				Direction:   "downward",
				SeatState:   "ASSIGNED",
				Expected:    "deny",
			},
			{
				Name:        "Cross-Domain (No Approval)",
				Description: "Cross-domain downward authority requires parent approval",
				Direction:   "downward",
				SeatState:   "OCCUPIED",
				Expected:    "deny (no approval)",
			},
			{
				Name:        "Frozen Seat",
				Description: "Frozen seat has no authority",
				Direction:   "upward",
				SeatState:   "FROZEN",
				Expected:    "deny",
			},
		},
	}

	writeJSON(w, preview)
}

// EvaluateFlow evaluates an authority flow decision (read-only, no side effects)
// POST /api/authority/flow/eval
func (h *TriadHandler) EvaluateFlow(w http.ResponseWriter, r *http.Request) {
	var input FlowEvalInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if input.IdentityID == "" {
		http.Error(w, "missing identity_id", http.StatusBadRequest)
		return
	}
	if input.Action == "" {
		http.Error(w, "missing action", http.StatusBadRequest)
		return
	}
	if input.Direction == "" {
		input.Direction = "upward" // default
	}

	// Fetch triad for identity
	triad, err := h.triadRepo.GetIdentityTriad(r.Context(), input.IdentityID)
	if err != nil {
		http.Error(w, "triad not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Build triad seat status
	triadSeats := make([]TriadSeatStatus, 0, 3)
	if triad.Terra != nil {
		triadSeats = append(triadSeats, TriadSeatStatus{
			Layer:  "terra",
			State:  triad.Terra.State,
			Frozen: triad.Terra.IsFrozen(),
		})
	}
	if triad.Numen != nil {
		triadSeats = append(triadSeats, TriadSeatStatus{
			Layer:  "numen",
			State:  triad.Numen.State,
			Frozen: triad.Numen.IsFrozen(),
		})
	}
	if triad.Lima != nil {
		triadSeats = append(triadSeats, TriadSeatStatus{
			Layer:  "lima",
			State:  triad.Lima.State,
			Frozen: triad.Lima.IsFrozen(),
		})
	}

	// Use flow engine (simple evaluation for GOV-2)
	seatState := "EMPTY"
	seatFrozen := false
	if triad.Lima != nil {
		seatState = triad.Lima.State
		seatFrozen = triad.Lima.IsFrozen()
	}

	result := authority.EvaluateAuthority(
		input.Action,
		seatState,
		authority.AuthorityDirection(input.Direction),
		input.ActionDomain,
		input.SeatDomain,
		input.ParentApproved,
		seatFrozen,
	)

	// Build response
	response := &FlowEvalResult{
		Allow:       result.Allow,
		Reason:      result.Reason,
		TriadSeats:  triadSeats,
		EvaluatedAt: GetTimestamp(),
		Details: map[string]any{
			"direction":       input.Direction,
			"action":          input.Action,
			"parent_approved": input.ParentApproved,
			"errors":          result.Errors,
		},
	}

	writeJSON(w, response)
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// GOV-3: Write endpoints for seat transitions

// TransitionSeat handles single seat state transition requests
// POST /api/authority/seat/transition
func (h *TriadHandler) TransitionSeat(w http.ResponseWriter, r *http.Request) {
	if h.mutationEngine == nil {
		http.Error(w, "mutation engine not configured", http.StatusServiceUnavailable)
		return
	}

	var req authority.SeatTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	result, err := h.mutationEngine.MutateSeat(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !result.OK {
		w.WriteHeader(http.StatusForbidden)
		writeJSON(w, result)
		return
	}

	writeJSON(w, result)
}

// TransitionSeatBatch handles batch seat state transition requests
// POST /api/authority/seat/transition/batch
func (h *TriadHandler) TransitionSeatBatch(w http.ResponseWriter, r *http.Request) {
	if h.mutationEngine == nil {
		http.Error(w, "mutation engine not configured", http.StatusServiceUnavailable)
		return
	}

	var req authority.SeatTransitionBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	result, err := h.mutationEngine.MutateSeatBatch(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, result)
}

// formatTime formats a time.Time to RFC3339 string
func formatTime(t any) string {
	return GetTimestamp() // Simplified - use actual time if available
}
