# DIS-PERSONAL v0.5 — Network Sovereignty

## Run (server mode)
```
go mod tidy
go run main.go --serve
# Server listens on api_host:api_port from config.yaml (defaults 0.0.0.0:8080)
```

## Endpoints
- `GET  /health`   → `{ "status": "ok", "version": "v0.5" }`
- `GET  /policy`   → returns active policy and checksum
- `GET  /receipts` → returns public receipts list (no identity_id)
- `POST /act`      → JSON: `{ "by": "domain.null", "scope": "identity.confirm", "nonce": "optional" }`

### Example (curl)
```
curl -X POST http://localhost:8080/act \
  -H 'Content-Type: application/json' \
  -d '{"by":"domain.null","scope":"identity.confirm"}'
```

## Privacy
- No identity UUIDs are exposed via the API.
- Receipts return: action, by, scope, timestamp, nonce, policy_checksum, signature, receipt_id.
- Policy content is only exposed via the explicit `/policy` endpoint.
