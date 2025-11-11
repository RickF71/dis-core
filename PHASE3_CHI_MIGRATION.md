# Phase 3 Chi Router Migration - DIS-Core API

This document describes the Phase 3 migration that implements chi router integration with format-aware endpoints as specified in `.copilot-instructions`.

## Migration Overview

The migration replaces the existing `http.ServeMux` based routing with `github.com/go-chi/chi/v5` router, adding format-aware endpoints that support multiple response formats through subpath detection.

## New Files Created

### Core Router Infrastructure
- `internal/api/chi_router.go` - Main chi router with comprehensive route registration
- `internal/api/chi_migration.go` - Migration utilities and server wrapper

### Format-Aware Handlers
- `internal/api/handlers_domain_chi.go` - Domain endpoints with JSON/file/text support
- `internal/api/handlers_policy_chi.go` - Policy management with format switching
- `internal/api/handlers_status_chi.go` - Status endpoints with text format support
- `internal/api/handlers_file_chi.go` - File operations with format awareness
- `internal/api/handlers_receipt_chi.go` - Receipt management with multiple formats

## Format Support

### Subpath Format Detection
Formats are detected via URL subpaths (no dot extensions):
- `/api/endpoint` - JSON format (default)
- `/api/endpoint/file` - Raw file/content format
- `/api/endpoint/text` - Plain text format
- `/api/endpoint/json` - Explicit JSON format

### Supported Formats by Endpoint Type
- **Domain endpoints**: JSON, File, Text
- **Policy endpoints**: File (raw rego), JSON (metadata), Text
- **Status endpoints**: JSON, Text
- **Receipt endpoints**: File (raw JSON), JSON (with metadata), Text
- **File endpoints**: JSON, File

## Integration Examples

### Basic Migration
```go
// Get existing server instance
server := getYourExistingAPIServer()

// Migrate to chi router
updatedServer := api.MigrateToChiRouter(server)

// Start chi-based server
log.Fatal(updatedServer.StartChiServer(":8080"))
```

### Advanced Integration with Existing main.go
```go
func main() {
    // Load your existing configuration
    cfg := loadConfiguration()

    // Initialize your app as usual
    app := initializeApp(cfg)
    defer app.Close()

    // Get the existing API server
    server := app.GetAPIServer()

    // Migrate to chi router
    updatedServer := api.MigrateToChiRouter(server)

    // Optional: Print all routes for debugging
    updatedServer.PrintRoutes()

    // Start with chi router
    log.Fatal(updatedServer.StartChiServer(":8080"))
}
```

## API Usage Examples

### Domain Operations
```bash
# Get domain as JSON (default)
curl http://localhost:8080/api/domain/example

# Get domain as raw file
curl http://localhost:8080/api/domain/example/file

# Get domain CSS as raw CSS
curl http://localhost:8080/api/domain/example/css/file

# Get domain list as text
curl http://localhost:8080/api/domains/text
```

### Policy Operations
```bash
# Get policy as raw rego file
curl http://localhost:8080/api/policy/auth/file

# Get policy with metadata as JSON
curl http://localhost:8080/api/policy/auth/json

# List policies as export file
curl http://localhost:8080/api/policy/list/file
```

### Status and Health
```bash
# Get status as JSON
curl http://localhost:8080/api/status

# Get status as formatted text
curl http://localhost:8080/api/status/text

# Check ping as text
curl http://localhost:8080/api/ping/text
```

### Receipt Operations
```bash
# Get receipt as raw JSON
curl http://localhost:8080/api/receipt/123/file

# Verify receipt with text output
curl http://localhost:8080/api/receipt/123/verify/text

# List receipts as text summary
curl http://localhost:8080/api/receipts/text
```

## Implementation Details

### Format Detection
The `DetectFormat(r *http.Request)` function examines the URL path for format indicators:
- Checks for `/file`, `/text`, `/json`, `/raw` suffixes
- Defaults to JSON format
- Handles format-specific error responses

### Response Helpers
Unified response helpers maintain consistency:
- `JSON(w, status, data)` - Standard JSON responses
- `JSONError(w, status, message)` - Error responses
- `JSONOk(w, message)` - Success responses
- `ServeAsFile(w, data)` - File downloads
- `ServeAsText(w, content)` - Plain text responses

### Route Registration
Format-aware routes use `RegisterFormatAwareRoute()`:
```go
s.RegisterFormatAwareRoute(r, "GET", "/api/domain/{id}",
    s.handleGetDomainFormatAware,
    []Format{FormatJSON, FormatFile, FormatText})
```

## Migration Checklist

- [x] Chi router infrastructure created
- [x] Format detection and response helpers implemented
- [x] Domain endpoints migrated with format support
- [x] Policy endpoints migrated with format support
- [x] Status endpoints migrated with text support
- [x] File operations migrated with format awareness
- [x] Receipt management migrated with multiple formats
- [x] Migration utilities and server wrapper created
- [ ] Integration with existing server startup code
- [ ] Testing and validation
- [ ] Documentation updates

## Next Steps

1. **Server Integration**: Update your main server startup to use `MigrateToChiRouter()`
2. **Business Logic Integration**: Replace placeholder methods with actual implementations
3. **Testing**: Validate format switching and endpoint functionality
4. **Documentation**: Update API documentation with new format capabilities
5. **Backward Compatibility**: Ensure existing clients continue to work

## Benefits

- **Format Flexibility**: Multiple response formats for the same data
- **Better UX**: Text formats for human-readable status, file formats for downloads
- **Modern Routing**: Chi router provides better performance and middleware support
- **Extensibility**: Easy to add new formats or endpoints
- **Consistency**: Unified response patterns across all endpoints

## Troubleshooting

### Common Issues
- Ensure chi router dependency is available: `go get github.com/go-chi/chi/v5`
- Format detection expects exact subpath matches (case sensitive)
- Placeholder methods need replacement with actual business logic
- Middleware order matters for format detection

### Debugging
- Use `updatedServer.PrintRoutes()` to see all registered routes
- Check format detection with different URL patterns
- Verify response headers for content types
- Test unsupported format handling
