package freeze

default deny = false

# Example freeze rule: deny all actions if domain is marked frozen
deny = true {
  input.domainVars.frozen == true
}
