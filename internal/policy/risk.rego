package risk

# -----------------------------------------------------------
# Risk Policy — assign simple quantitative or qualitative score
# -----------------------------------------------------------

default score = 0.0

# Example: domain.freeze.override.v1 is a higher-risk operation
score = 0.9 {
  input.action == "domain.freeze.override.v1"
}

details = {
  "score": score,
  "note": note
}

note = msg {
  score == 0.0
  msg := "baseline risk"
}

note = msg {
  score > 0.0
  msg := sprintf("elevated risk for action: %s", [input.action])
}
