package gates

import data.dis.seats

# -----------------------------------------------------------
# Gates Policy — Base allow/deny logic for DIS actions
# Phase S2 Integration: Include seat authorization
# -----------------------------------------------------------

# Default: all actions allowed unless overridden by domain rule
default allow = true

# Set of blocked actions (core deny set)
# Represented as predicate form so rules can be dynamic.
deny_action["domain.freeze.override.v1"] { true }
deny_action["ci.call.test.v1"] { input.payload.block == true }

# Example narrow test rule (AT-1): block test CI call when payload.block == true
deny_test_block {
  input.action == "ci.call.test.v1"
  p := input.payload
  p.block == true
}

# Allow if the action is NOT in the deny set, not blocked by test rule, AND seats allow
allow {
  not deny_action[input.action]
  not deny_test_block
  seats.allow
}

rule = "deny-test-block" { deny_test_block }
rule = "default-allow" { not deny_test_block }

policy_ref = "ci_rules:ci_call_test_block_v1" { deny_test_block }
policy_ref = "gates:default" { not deny_test_block }

# Standardized deny code for AT-1 rule so runtimes can produce stable reason codes
deny_code = "at1.ci_call_test_block_v1" { deny_test_block }

details = {
  "allow": allow,
  "checked": true,
  "rule": rule,
  "deny_code": deny_code,
  "seat_authorized": seats.allow,
  "policy_ref": policy_ref,
}
