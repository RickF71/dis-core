# Lima Policy: Consent and Authority Layer (GOV-1 Layer 3)
# Purpose: Ensures consent and authority flow compliance
# Evaluation Order: Third layer (after terra + numen)

package dis.lima

import future.keywords.if
import data.dis.terra
import data.dis.numen

# Default deny-by-default security model
default allow := false
default deny := false

# Lima seat must be valid for authority actions
lima_seat_valid if {
	input.lima_seat_state
	input.lima_seat_state != "EMPTY"
	input.lima_seat_state != "FROZEN"
}

# Upward authority allowed if lima seat is ASSIGNED or OCCUPIED
upward_authority if {
	input.direction == "upward"
	lima_seat_valid
}

upward_authority if {
	input.direction == "upward"
	input.lima_seat_state == "ASSIGNED"
}

upward_authority if {
	input.direction == "upward"
	input.lima_seat_state == "OCCUPIED"
}

# Downward authority requires OCCUPIED + not frozen + approval
downward_authority if {
	input.direction == "downward"
	input.lima_seat_state == "OCCUPIED"
	not input.seat_frozen
	downward_approved
}

# Downward approval logic
downward_approved if {
	# Within-domain actions allowed
	input.action_domain
	input.seat_domain
	input.action_domain == input.seat_domain
}

downward_approved if {
	# Cross-domain with parent approval
	input.action_domain
	input.seat_domain
	input.action_domain != input.seat_domain
	input.parent_approved == true
}

# Lateral authority requires OCCUPIED in same domain
lateral_authority if {
	input.direction == "lateral"
	input.lima_seat_state == "OCCUPIED"
	input.action_domain == input.seat_domain
}

# Allow if terra + numen allow AND lima authority satisfied
allow if {
	terra.export_allow
	numen.export_allow
	lima_seat_valid
	upward_authority
}

allow if {
	terra.export_allow
	numen.export_allow
	lima_seat_valid
	downward_authority
}

allow if {
	terra.export_allow
	numen.export_allow
	lima_seat_valid
	lateral_authority
}

# Deny if terra or numen deny (cannot override parent layers)
deny if {
	terra.export_deny
}

deny if {
	numen.export_deny
}

# Deny if lima seat is invalid
deny if {
	not lima_seat_valid
}

# Deny if seat is frozen
deny if {
	input.seat_frozen == true
}

# Deny if downward authority attempted without approval
deny if {
	input.direction == "downward"
	not downward_authority
}

# Deny if lateral authority attempted cross-domain
deny if {
	input.direction == "lateral"
	input.action_domain != input.seat_domain
}

# Export evaluation results (required for policy chaining)
export_allow := allow
export_deny := deny
export_reason := reason

# Reason messages for debugging/auditing
reason := "lima.member upward authority satisfied" if {
	allow
	upward_authority
}

reason := "lima.member downward authority satisfied" if {
	allow
	downward_authority
}

reason := "lima.member lateral authority satisfied" if {
	allow
	lateral_authority
}

reason := "terra layer denied" if {
	deny
	terra.export_deny
}

reason := "numen layer denied" if {
	deny
	numen.export_deny
}

reason := "lima.member seat invalid or frozen" if {
	deny
	not lima_seat_valid
}

reason := "seat is frozen" if {
	deny
	input.seat_frozen == true
}

reason := "downward authority requires parent approval" if {
	deny
	input.direction == "downward"
	not downward_authority
}

reason := "lateral authority not allowed cross-domain" if {
	deny
	input.direction == "lateral"
	input.action_domain != input.seat_domain
}
