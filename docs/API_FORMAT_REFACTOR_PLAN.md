# DIS-Core API Format Refactor Implementation Plan

## Phase 1: Foundation - JSON Helper Standardization

### 1.1 Extend JSON Response Helpers
**File**: `internal/api/respond.go`

Add missing JSON helper functions:
```go
// JSONError writes a standardized error response
func JSONError(w http.ResponseWriter, status int, message string) {
    JSON(w, status, map[string]any{
        "error":  message,
        "status": status,
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    })
}

// JSONUnsupportedFormat responds with format not supported error
func JSONUnsupportedFormat(w http.ResponseWriter, format string) {
    JSONError(w, http.StatusNotAcceptable,
        fmt.Sprintf("unsupported format: %s", format))
}

// JSONSuccess writes a standardized success response
func JSONSuccess(w http.ResponseWriter, message string, data any) {
    JSON(w, http.StatusOK, map[string]any{
        "success": true,
        "message": message,
        "data": data,
    })
}
```

### 1.2 Replace http.Error Calls (95% of endpoints)

**Priority Order**:
1. Core domain handlers (domain.go, domain_get.go, etc.)
2. Policy and file handlers
3. Status and utility endpoints
4. Not-implemented handlers

**Pattern Replacement**:
```go
// Before:
http.Error(w, "domain not found", http.StatusNotFound)

// After:
JSONError(w, http.StatusNotFound, "domain not found")
```

**Files to update**:
- `domain.go` (10+ instances)
- `domain_get.go` (5+ instances)
- `domain_list.go` (5+ instances)
- `domain_update.go` (3+ instances)
- `routes_policy_files.go` (8+ instances)
- `routes_domain_files.go` (10+ instances)
- And 20+ other files

### 1.3 Standardize JSON Encoding

Replace manual `json.NewEncoder(w).Encode()` with `JSON()` helper:

**Files to update**:
- `status.go`
- `domain_create.go`
- `routes_authority.go`
- `domain_policy.go`
- And 15+ other files

## Phase 2: Format Detection Infrastructure

### 2.1 Add Format Detection Function
**File**: `internal/api/format.go` (new)

```go
package api

import (
    "strings"
    "net/http"
)

type Format string

const (
    FormatJSON      Format = "json"
    FormatFile      Format = "file"
    FormatText      Format = "text"
    FormatCBOR      Format = "cbor"
    FormatEncrypted Format = "encrypted"
)

// DetectFormat extracts format from URL path suffix
func DetectFormat(r *http.Request) Format {
    path := r.URL.Path

    if strings.HasSuffix(path, "/json") { return FormatJSON }
    if strings.HasSuffix(path, "/file") { return FormatFile }
    if strings.HasSuffix(path, "/text") { return FormatText }
    if strings.HasSuffix(path, "/cbor") { return FormatCBOR }
    if strings.HasSuffix(path, "/encrypted") { return FormatEncrypted }

    // Default to JSON for backward compatibility
    return FormatJSON
}

// StripFormatSuffix removes format suffix from path
func StripFormatSuffix(path string) string {
    suffixes := []string{"/json", "/file", "/text", "/cbor", "/encrypted"}
    for _, suffix := range suffixes {
        if strings.HasSuffix(path, suffix) {
            return strings.TrimSuffix(path, suffix)
        }
    }
    return path
}
```

### 2.2 Add Route Registration for Format Suffixes
**File**: `internal/api/routes.go`

Add format-aware route registration:
```go
// RegisterFormatAwareRoute registers a handler for multiple format suffixes
func (s *Server) RegisterFormatAwareRoute(mux *http.ServeMux, pattern, handler string, formats []Format) {
    basePattern := pattern

    // Register base pattern (defaults to JSON)
    mux.HandleFunc(pattern, handler)

    // Register format-specific patterns
    for _, format := range formats {
        formatPattern := pattern + "/" + string(format)
        mux.HandleFunc(formatPattern, handler)
    }
}
```

### 2.3 Add Format-Specific Response Functions
**File**: `internal/api/respond.go`

```go
// RespondWithFormat handles format-specific responses
func RespondWithFormat(w http.ResponseWriter, r *http.Request, data any) {
    format := DetectFormat(r)

    switch format {
    case FormatJSON:
        JSON(w, http.StatusOK, data)
    case FormatFile:
        ServeAsFile(w, data)
    case FormatText:
        ServeAsText(w, data)
    case FormatCBOR:
        ServeCBOR(w, data)
    case FormatEncrypted:
        ServeEncrypted(w, data)
    default:
        JSONUnsupportedFormat(w, string(format))
    }
}

// ServeAsFile serves data as downloadable file
func ServeAsFile(w http.ResponseWriter, data any) {
    // Implementation for file download
    w.Header().Set("Content-Type", "application/octet-stream")
    w.Header().Set("Content-Disposition", "attachment")
    // Marshal and serve data
}

// ServeAsText serves data as plain text
func ServeAsText(w http.ResponseWriter, data any) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    // Convert data to text representation
}

// ServeCBOR serves data in CBOR format
func ServeCBOR(w http.ResponseWriter, data any) {
    // CBOR implementation
    JSONUnsupportedFormat(w, "cbor") // Placeholder
}

// ServeEncrypted serves encrypted data
func ServeEncrypted(w http.ResponseWriter, data any) {
    // Encryption implementation
    JSONUnsupportedFormat(w, "encrypted") // Placeholder
}
```

## Phase 3: Handler Migration Strategy

### 3.1 Data Endpoints (Priority 1)

**Target handlers**:
- `handleGetDomain` - domain data with /json and /file support
- `handleDomainList` - domain list with /json and /file (export) support
- `HandleStatus` - status with /json and /text support

**Implementation pattern**:
```go
func (s *Server) handleGetDomain(w http.ResponseWriter, r *http.Request) {
    format := DetectFormat(r)

    // Strip format suffix to get actual domain ID
    path := StripFormatSuffix(r.URL.Path)
    id := extractDomainID(path)

    if id == "" {
        switch format {
        case FormatJSON:
            JSONError(w, http.StatusBadRequest, "missing domain id")
        default:
            JSONUnsupportedFormat(w, string(format))
        }
        return
    }

    // Get domain data
    domain, err := s.getDomain(ctx, id)
    if err != nil {
        switch format {
        case FormatJSON:
            JSONError(w, http.StatusNotFound, "domain not found")
        default:
            JSONUnsupportedFormat(w, string(format))
        }
        return
    }

    // Format-specific response
    switch format {
    case FormatJSON:
        JSON(w, http.StatusOK, domain)
    case FormatFile:
        ServeAsFile(w, domain) // Download as JSON file
    default:
        JSONUnsupportedFormat(w, string(format))
    }
}
```

### 3.2 File Endpoints (Priority 2)

**Target handlers**:
- `handleDomainCSS` - CSS content with /file (raw) and /json (metadata) support
- `handleDomainFileGet` - file content with /file and /json support
- `handleGetPolicy` - policy content with /file and /json support

**Implementation pattern**:
```go
func (s *Server) handleDomainCSS(w http.ResponseWriter, r *http.Request) {
    format := DetectFormat(r)
    id := r.PathValue("id")

    css, err := s.getDomainCSS(ctx, id)
    if err != nil {
        switch format {
        case FormatJSON:
            JSONError(w, http.StatusNotFound, "domain not found")
        default:
            JSONUnsupportedFormat(w, string(format))
        }
        return
    }

    switch format {
    case FormatFile:
        w.Header().Set("Content-Type", "text/css; charset=utf-8")
        io.WriteString(w, css.Content)
    case FormatJSON:
        JSON(w, http.StatusOK, map[string]any{
            "domain_id": id,
            "css_content": css.Content,
            "last_modified": css.UpdatedAt,
            "size": len(css.Content),
        })
    default:
        JSONUnsupportedFormat(w, string(format))
    }
}
```

### 3.3 Status/Utility Endpoints (Priority 3)

**Target handlers**:
- `handlePing` - simple ping with /json and /text support
- `handleHealth` - health check with /json and /text support
- `handleDomainAnnounce` - announcements with /json and /text support

## Phase 4: Route Registration Updates

### 4.1 Update routes.go with Format Support

```go
// Replace existing route registrations
func (s *Server) registerRoutes(mux *http.ServeMux) {
    // Status endpoints - support json and text
    s.RegisterFormatAwareRoute(mux, "GET /api/ping", s.handlePing,
        []Format{FormatJSON, FormatText})
    s.RegisterFormatAwareRoute(mux, "GET /api/status", s.HandleStatus,
        []Format{FormatJSON, FormatText})

    // Domain endpoints - support json and file
    s.RegisterFormatAwareRoute(mux, "GET /api/domain/{id}", s.handleGetDomain,
        []Format{FormatJSON, FormatFile})
    s.RegisterFormatAwareRoute(mux, "GET /api/domains", s.handleDomainList,
        []Format{FormatJSON, FormatFile})

    // CSS endpoint - support file and json
    s.RegisterFormatAwareRoute(mux, "GET /api/domain/{id}/css", s.handleDomainCSS,
        []Format{FormatFile, FormatJSON})

    // Continue for all endpoints...
}
```

### 4.2 Backward Compatibility

Ensure existing routes without format suffixes continue to work:
- Default format is JSON for data endpoints
- Default format is file for file content endpoints
- All existing URLs remain functional

## Implementation Timeline

### Week 1: Foundation
- [ ] Add JSON helper functions to respond.go
- [ ] Create format.go with detection logic
- [ ] Update 5 core domain handlers with new error patterns

### Week 2: Error Standardization
- [ ] Replace http.Error in all domain handlers (20+ files)
- [ ] Replace http.Error in policy handlers (5 files)
- [ ] Replace http.Error in file handlers (5 files)

### Week 3: JSON Standardization
- [ ] Replace manual json.NewEncoder with JSON helper (15+ files)
- [ ] Add format detection to route handlers
- [ ] Implement ServeAsFile and ServeAsText functions

### Week 4: Format Migration
- [ ] Migrate core data endpoints (domain, status, lists)
- [ ] Migrate file content endpoints (CSS, policies, files)
- [ ] Update route registrations with format support

### Week 5: Testing & Documentation
- [ ] Test all endpoints with format suffixes
- [ ] Update API documentation
- [ ] Add integration tests for format switching

## Testing Strategy

### Unit Tests
```go
func TestFormatDetection(t *testing.T) {
    tests := []struct {
        path string
        expected Format
    }{
        {"/api/domain/123", FormatJSON},
        {"/api/domain/123/json", FormatJSON},
        {"/api/domain/123/file", FormatFile},
        {"/api/domain/123/text", FormatText},
    }
    // Test implementation...
}
```

### Integration Tests
```go
func TestDomainEndpointFormats(t *testing.T) {
    // Test /api/domain/123 (JSON default)
    // Test /api/domain/123/json (explicit JSON)
    // Test /api/domain/123/file (file download)
    // Test /api/domain/123/invalid (error response)
}
```

## Risk Mitigation

### Backward Compatibility
- All existing routes remain functional
- Default format behavior unchanged
- Gradual migration allows rollback at any stage

### Performance Impact
- Format detection is simple string operations
- No significant performance overhead
- Response caching can be added per format

### Error Handling
- Consistent error responses across all formats
- Proper HTTP status codes maintained
- Enhanced error information in JSON format

## Success Metrics

### Code Quality
- ✅ 95% reduction in http.Error usage
- ✅ 100% of endpoints use JSON helpers for JSON responses
- ✅ Consistent error response format across all endpoints

### API Consistency
- ✅ All data endpoints support /json and /file formats
- ✅ All file endpoints support /file and /json formats
- ✅ All status endpoints support /json and /text formats

### Developer Experience
- ✅ Clear format suffix conventions
- ✅ Standardized response patterns
- ✅ Improved error messages and debugging

---
*Implementation plan created: November 10, 2025*
