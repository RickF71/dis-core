# MOAR-CORS v1 Implementation

## Overview
MOAR-CORS v1 is a dynamic CORS middleware that validates origins against a whitelist and supports environment-based configuration for production deployments.

## Features

### 1. **Dynamic Origin Checking**
Instead of hardcoding a single origin, MOAR-CORS checks incoming requests against:
- Built-in dev origins (localhost:5173, 127.0.0.1:5173, [::1]:5173)
- Environment-configured origins (via `DIS_ALLOWED_ORIGINS`)

### 2. **Environment Variable Support**
Set `DIS_ALLOWED_ORIGINS` to a comma-separated list of allowed origins:
```bash
export DIS_ALLOWED_ORIGINS='https://finagler.dis.example,https://app.dis.example'
./dis-core-moar-cors
```

### 3. **Complete Header Set**
When an origin is allowed, MOAR-CORS sets:
- `Access-Control-Allow-Origin: <validated-origin>`
- `Access-Control-Allow-Credentials: true`
- `Access-Control-Allow-Headers: Content-Type, Authorization, X-External-User, X-Acting-As, X-Requested-With, Accept`
- `Access-Control-Allow-Methods: GET,POST,OPTIONS`
- `Vary: Origin` (cache safety)

### 4. **Automatic Preflight Handling**
OPTIONS requests are automatically handled and return 200 OK with CORS headers.

## Architecture

### File Structure
```
internal/cors/middleware.go    - MOAR-CORS v1 implementation
internal/api/middleware/chain.go - Updated to use cors.Middleware
```

### Integration
MOAR-CORS is applied globally as the **first middleware** in the chain:
```go
func Attach(r *chi.Mux, pool *pgxpool.Pool, engine policy.PolicyEngine) {
    // CORS middleware (MUST be first to handle OPTIONS preflight)
    r.Use(cors.Middleware)

    // Chi built-in middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    // ... rest of middleware stack
}
```

## Security

### Origin Validation
- **Whitelist-based**: Only explicitly allowed origins receive CORS headers
- **No wildcards**: Never uses `Access-Control-Allow-Origin: *` with credentials
- **Strict matching**: Origin must match exactly (protocol, host, port)

### Headers Allowed
All headers required by Finagler and future features:
- `Content-Type` - JSON POST requests
- `Authorization` - Token-based auth
- `X-External-User` - Actor pipeline (dev mode)
- `X-Acting-As` - Seat switching
- `X-Requested-With` - Fetch/XHR compatibility
- `Accept` - Content negotiation

## Testing

### Run Tests
```bash
# Start dis-core with MOAR-CORS
cd /home/rick/dev/DIS/dis-core
source .env.postgres
./dis-core-moar-cors

# In another terminal, run tests
./test-moar-cors.sh
```

### Expected Results
- ✅ Allowed origins receive full CORS headers
- ✅ Disallowed origins receive NO CORS headers
- ✅ OPTIONS preflight returns 200 OK
- ✅ All required headers present

### Manual Testing
```bash
# Test allowed origin
curl -I http://localhost:8080/api/status -H "Origin: http://localhost:5173"

# Test blocked origin
curl -I http://localhost:8080/api/status -H "Origin: http://evil.com"
```

## Migration from Old CORS

### Before (internal/api/middleware/cors.go)
```go
// Hardcoded single origin
w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
```

### After (internal/cors/middleware.go)
```go
// Dynamic origin checking
origin := r.Header.Get("Origin")
if isAllowedOrigin(origin) {
    w.Header().Set("Access-Control-Allow-Origin", origin)
    // ... other headers
}
```

## Production Configuration

### Development
No configuration needed. Built-in dev origins are automatically allowed:
- http://localhost:5173
- http://127.0.0.1:5173
- http://[::1]:5173

### Production
Set environment variable before starting dis-core:
```bash
export DIS_ALLOWED_ORIGINS='https://finagler.production.com,https://app.production.com'
./dis-core-moar-cors
```

### Docker/Kubernetes
```yaml
env:
  - name: DIS_ALLOWED_ORIGINS
    value: "https://finagler.production.com,https://app.production.com"
```

## Troubleshooting

### CORS Errors in Browser
1. Check browser console for exact error
2. Verify origin is in allowed list: `echo $DIS_ALLOWED_ORIGINS`
3. Run diagnostics: `./dis-cors-diagnostics.sh`
4. Check server logs for incoming Origin header

### No CORS Headers
- Verify dis-core is using `dis-core-moar-cors` binary
- Check middleware order (CORS must be first)
- Verify Origin header is sent by browser

### Credentials Not Working
- Ensure `Access-Control-Allow-Credentials: true` is present
- Verify origin is NOT wildcard (*)
- Check cookie settings (SameSite, Secure flags)

## Development Notes

### Adding New Origins
1. Development: Add to `allowedDevOrigins` in `middleware.go`
2. Production: Add to `DIS_ALLOWED_ORIGINS` environment variable

### Adding New Headers
Update the `Access-Control-Allow-Headers` line in `Middleware()`:
```go
w.Header().Set("Access-Control-Allow-Headers",
    "Content-Type, Authorization, X-External-User, X-Acting-As, X-Requested-With, Accept, X-New-Header")
```

### Debugging
Add logging to `isAllowedOrigin()`:
```go
func isAllowedOrigin(origin string) bool {
    log.Printf("[CORS] Checking origin: %s", origin)
    // ... validation logic
}
```

## Future Enhancements

- [ ] Wildcard subdomain support (*.dis.example)
- [ ] Per-route CORS policies
- [ ] Rate limiting by origin
- [ ] CORS metrics and monitoring
- [ ] Dynamic origin loading from database
- [ ] JWT-based origin validation

## Related Files

- `internal/cors/middleware.go` - MOAR-CORS v1 implementation
- `internal/api/middleware/chain.go` - Middleware stack configuration
- `internal/api/dev.go` - Dev auth endpoints
- `test-moar-cors.sh` - CORS testing script
- `dis-cors-diagnostics.sh` - Comprehensive CORS diagnostics

## Changelog

### v1.0.0 (2025-11-17)
- Initial MOAR-CORS v1 implementation
- Dynamic origin checking with whitelist
- Environment variable support (DIS_ALLOWED_ORIGINS)
- Complete header set for Finagler integration
- Automatic OPTIONS preflight handling
- Replaces hardcoded origin in old middleware
