package freeze

# -----------------------------------------------------------
# Freeze Policy — controls frozen/unfrozen state for a domain
# -----------------------------------------------------------

default active = false

# Example: if the domain contains "sandbox", it is auto-frozen
active {
  contains(input.domain, "sandbox")
}

details = {
  "frozen": active,
  "reason": reason
}

reason = msg {
  not active
  msg := "domain is active and not frozen"
}

reason = msg {
  active
  msg := sprintf("domain %s is currently frozen (sandbox rule)", [input.domain])
}
