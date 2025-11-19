# GOV-3: Seat transition policy (dis.seat)
# Governs allowed state transitions for terra/numen/lima seats
# Enforces transition matrix, actor consent, permissions, and parent decisions

package dis.seat

default allow := false
default reason := "denied by default"

# ============================================================================
# Validation Rules
# ============================================================================

valid_layer(l) {
	l == "terra"
}

valid_layer(l) {
	l == "numen"
}

valid_layer(l) {
	l == "lima"
}

valid_state(s) {
	s == "EMPTY"
}

valid_state(s) {
	s == "ASSIGNED"
}

valid_state(s) {
	s == "OCCUPIED"
}

valid_state(s) {
	s == "FROZEN"
}

# ============================================================================
# Parent Decision Override
# ============================================================================

# Parent layer cannot be overridden - if parent denied, stop here
parent_blocked {
	input.context.parent_decision == "deny"
}

# ============================================================================
# Transition Matrix
# ============================================================================

# Define allowed state transitions
allowed_transition(from, to) {
	from == "EMPTY"
	to == "ASSIGNED"
}

allowed_transition(from, to) {
	from == "ASSIGNED"
	to == "OCCUPIED"
}

allowed_transition(from, to) {
	from == "OCCUPIED"
	to == "FROZEN"
}

# Thaw: Frozen back to ASSIGNED (break-glass scenario)
allowed_transition(from, to) {
	from == "FROZEN"
	to == "ASSIGNED"
}

# ============================================================================
# Actor Consent Requirements
# ============================================================================

# Actor must have lima seat in OCCUPIED state to authorize mutations
actor_ok {
	some i
	input.actor.triad_seats[i].layer == "lima"
	input.actor.triad_seats[i].state == "OCCUPIED"
}

# ============================================================================
# Permission Checks
# ============================================================================

# Require "seat.transition" permission in context
perm_ok {
	input.context.permissions[_] == "seat.transition"
}

# ============================================================================
# Freeze Policy Hook
# ============================================================================

# Check if domain freeze policy denies this action
not_frozen {
	not input.context.freeze_denied
}

# ============================================================================
# Allow Decision
# ============================================================================

allow {
	not parent_blocked
	valid_layer(input.request.layer)
	valid_state(input.request.from)
	valid_state(input.request.to)
	allowed_transition(input.request.from, input.request.to)
	actor_ok
	perm_ok
	not_frozen
}

# ============================================================================
# Reason Generation
# ============================================================================

reason := "parent decision denied" {
	parent_blocked
}

reason := "invalid layer" {
	not valid_layer(input.request.layer)
}

reason := "invalid state" {
	not valid_state(input.request.from)
	not valid_state(input.request.to)
}

reason := "invalid transition" {
	not allowed_transition(input.request.from, input.request.to)
}

reason := "actor consent missing - lima seat must be OCCUPIED" {
	not actor_ok
}

reason := "permission missing - requires seat.transition" {
	not perm_ok
}

reason := "domain is frozen" {
	input.context.freeze_denied
}

reason := "allowed" {
	allow
}

# ============================================================================
# Exports for authority engine
# ============================================================================

export_allow := allow
export_deny := not allow
export_reason := reason
