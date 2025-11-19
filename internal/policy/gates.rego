package gates

import data.dis.seats

# -----------------------------------------------------------
# Gates Policy — Base allow/deny logic for DIS actions
# Phase S2 Integration: Include seat authorization
# -----------------------------------------------------------

# Default: all actions allowed unless overridden by domain rule
default allow = true

# Set of blocked actions
deny_actions = {"domain.freeze.override.v1"}

# Allow if the action is NOT in the deny set AND seats allow
allow {
  not deny_actions[input.action]
  seats.allow
}

details = {
  "allow": allow,
  "checked": true,
  "rule": "default-allow",
  "seat_authorized": seats.allow
}
