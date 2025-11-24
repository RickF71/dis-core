package freeze

# -----------------------------------------------------------
# Freeze Policy — controls frozen/unfrozen state for a domain
# -----------------------------------------------------------

default active = false

# Example: if the domain contains "sandbox", it is auto-frozen
active {
  contains(input.domain, "sandbox")
}

# AT-1 narrow rule: freeze when a test CI call explicitly requests a block
active {
  input.action == "ci.call.test.v1"
  p := input.payload
  p.block == true
}

details = {
  "frozen": active,
  "reason": reason,
  "policy_ref": policy_ref
}

reason = msg {
  not active
  msg := "domain is active and not frozen"
}

reason = msg {
  active
  msg := sprintf("domain %s is currently frozen (sandbox or test-block rule)", [input.domain])
}

policy_ref = "freeze:sandbox" { contains(input.domain, "sandbox") }
policy_ref = "ci_rules:ci_call_test_block_v1" { input.action == "ci.call.test.v1"; input.payload.block == true }
