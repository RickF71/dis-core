# Numen Policy: Meaning Layer (GOV-1 Layer 2)
# Purpose: Ensures semantic validity (schema compliance, canonical slots)
# Evaluation Order: Second layer (after terra)

package dis.numen

import future.keywords.if
import data.dis.terra

# Default deny-by-default security model
default allow := false
default deny := false

# Schema must be valid (schema_ref must exist)
schema_valid if {
	input.schema_ref
	input.schema_ref != ""
}

# Canonical greedy slots must exist (GOV-1 requirement)
canonical_valid if {
	input.payload
	input.payload.canonical_greedy
}

# If no payload, no canonical check needed
canonical_valid if {
	not input.payload
}

# Numen seat must not be EMPTY
numen_seat_valid if {
	input.numen_seat_state
	input.numen_seat_state != "EMPTY"
	input.numen_seat_state != "FROZEN"
}

# Allow if terra allows AND numen checks pass
allow if {
	terra.export_allow
	numen_seat_valid
	schema_valid
	canonical_valid
}

# Deny if terra denies (cannot override parent layer)
deny if {
	terra.export_deny
}

# Deny if numen seat is invalid
deny if {
	not numen_seat_valid
}

# Deny if schema is invalid
deny if {
	not schema_valid
}

# Deny if canonical_greedy is missing from payload
deny if {
	input.payload
	not input.payload.canonical_greedy
}

# Export evaluation results (required for policy chaining)
export_allow := allow
export_deny := deny
export_reason := reason

# Reason messages for debugging/auditing
reason := "numen.member meaning constraints satisfied" if {
	allow
}

reason := "terra layer denied" if {
	deny
	terra.export_deny
}

reason := "numen.member seat invalid or frozen" if {
	deny
	not numen_seat_valid
}

reason := "schema_ref missing or invalid" if {
	deny
	not schema_valid
}

reason := "canonical_greedy field missing from payload" if {
	deny
	input.payload
	not input.payload.canonical_greedy
}
