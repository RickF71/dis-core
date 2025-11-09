package gates

# -----------------------------------------------------------
# Gates Policy — Base allow/deny logic for DIS actions
# -----------------------------------------------------------

# Default: all actions allowed unless overridden by domain rule
default allow = true

# Set of blocked actions
deny_actions = {"domain.freeze.override.v1"}

# Allow if the action is NOT in the deny set
allow {
  not deny_actions[input.action]
}

details = {
  "allow": allow,
  "checked": true,
  "rule": "default-allow"
}
