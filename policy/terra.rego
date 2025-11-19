# Terra Policy: Existence Layer (GOV-1 Layer 1)
# Purpose: Ensures existential validity (identity exists, terra seat bound)
# Evaluation Order: First layer (foundation)

package dis.terra

import future.keywords.if

# Default deny-by-default security model
default allow := false
default deny := false

# Person exists if they have a valid terra.member binding
person_exists if {
	input.identity_id
	input.terra_seat_state
	input.terra_seat_state != "EMPTY"
}

# Allow if person exists and terra seat is valid
allow if {
	person_exists
	input.terra_seat_state == "ASSIGNED"
}

allow if {
	person_exists
	input.terra_seat_state == "OCCUPIED"
}

# Deny if identity_id is missing
deny if {
	not input.identity_id
}

# Deny if terra seat is EMPTY
deny if {
	input.terra_seat_state == "EMPTY"
}

# Deny if terra seat state is missing
deny if {
	not input.terra_seat_state
}

# Deny if terra seat is frozen
deny if {
	input.terra_seat_state == "FROZEN"
}

# Export evaluation results (required for policy chaining)
export_allow := allow
export_deny := deny
export_reason := reason

# Reason messages for debugging/auditing
reason := "terra.member seat verified (ASSIGNED)" if {
	allow
	input.terra_seat_state == "ASSIGNED"
}

reason := "terra.member seat verified (OCCUPIED)" if {
	allow
	input.terra_seat_state == "OCCUPIED"
}

reason := "terra.member seat missing or empty" if {
	deny
	not input.identity_id
}

reason := "terra.member seat is EMPTY" if {
	deny
	input.terra_seat_state == "EMPTY"
}

reason := "terra.member seat is FROZEN" if {
	deny
	input.terra_seat_state == "FROZEN"
}

reason := "terra.member seat state unknown" if {
	deny
	not input.terra_seat_state
}
