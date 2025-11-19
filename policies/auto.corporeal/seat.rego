# GOV-6: Auto-Corporeal Synthetic Domain Policy
# Deterministic consent simulator for testing

package dis.seat

# Default deny (require explicit token)
default allow = false
default reason = "no auto-token provided"

# Policy metadata
ref = "auto.corporeal/seat"
notify = true
message = "Notice: This domain is synthetic and has no corporeal agency."

# Auto-affirm condition
allow {
	input.context.note
	contains(input.context.note, "[always-affirm]")
}

reason = "auto-affirm: deterministic consent granted" {
	input.context.note
	contains(input.context.note, "[always-affirm]")
}

# Auto-deny condition
deny {
	input.context.note
	contains(input.context.note, "[always-deny]")
}

reason = "auto-deny: deterministic consent refused" {
	input.context.note
	contains(input.context.note, "[always-deny]")
}

# Fallback reason
reason = "missing deterministic token ([always-affirm] or [always-deny])" {
	not input.context.note
}

reason = "missing deterministic token ([always-affirm] or [always-deny])" {
	input.context.note
	not contains(input.context.note, "[always-affirm]")
	not contains(input.context.note, "[always-deny]")
}

# Synthetic domain declaration
synthetic = true

# Capabilities (all restricted for sandbox safety)
capabilities = {
	"delegation": false,
	"freeze_override": false,
	"break_glass": false,
	"authorize": false
}

# Corporeal warning
corporeal_warning = "This domain is synthetic and does not represent a true corporeal entity. It is designed for deterministic testing only."
