package dis.policy.at1_example

default allow = false

# Example rule: block operations when attribute "corrupt" is true
allow {
    input.context.attrs.corrupt == false
}
