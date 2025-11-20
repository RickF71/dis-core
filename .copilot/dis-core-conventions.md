# DIS-Core Conventions for Copilot

- Domain structs are canonical state holders
- Domain loader resolves governance
- Receipts represent all state mutations
- Freeze state = set-based + TTL + breakglass
- Policy evaluation happens through inherited rego
- Identities bind through domain rules, not global logic
- The database is not a source of truth; receipts are
