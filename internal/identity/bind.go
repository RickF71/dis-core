package identity

// copilot: Implement Identity Binding handler.
// POST /api/identity/bind: decode uid, domain, key; insert into identities table; emit ci.call.v1 receipt using Ledger.Record.
