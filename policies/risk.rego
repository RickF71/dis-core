package risk

default score = 0.1

# Simple risk heuristic
score = s {
  base := 0.1
  input.event.action == "domain.unfreeze.v1"
  s := base + 0.5
}
