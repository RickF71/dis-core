# DIS-Core API Endpoint Format Mapping Investigation

## Executive Summary

Investigation of 40+ API endpoints across the DIS-Core backend to prepare for unified format suffix refactor (/{format} pattern). The analysis reveals significant opportunities for standardization and improved error handling.

## Endpoint Format Analysis Table

| Route | Method | Handler | Current Formats | Uses JSON Helpers | Error Handling | Suggested Formats | File Path |
|-------|--------|---------|-----------------|-------------------|----------------|-------------------|-----------|
| `/api/ping` | GET | handlePing | text/plain | ❌ | http.Error | json, text | routes.go |
| `/api/health` | GET | handleHealth | text/plain | ❌ | http.Error | json, text | routes.go |
| `/api/status` | GET | HandleStatus | application/json | ❌ | manual JSON | json, text | status.go |
| `/api/authority/status` | GET | handleAuthorityStatus | application/json | ❌ | manual JSON | json, text | authority_status.go |
| `/api/policy/reload` | POST | handlePolicyReload | text/plain | ❌ | http.Error | json, text | routes.go |
| `/api/identity/bind` | POST | identity.BindHandler | not implemented | ❌ | http.Error | json | routes.go |
| `/api/domain` | POST | handleCreateDomain | application/json | ❌ | http.Error | json | domain_create.go |
| `/api/domain/{id}` | GET | handleGetDomain | application/json | ❌ | http.Error | json, file | domain_get.go |
| `/api/domains` | GET | handleDomainList | application/json | ❌ | http.Error | json, file | domain_list.go |
| `/api/domain` | GET | handleListDomains | application/json | ❌ | http.Error | json, file | domain_list.go |
| `/api/domain/{id}` | PUT | handleUpdateDomain | text/plain | ❌ | http.Error | json | domain_update.go |
| `/api/domain/default` | GET | handleGetDefaultDomain | application/json | ❌ | http.Error | json, file | domain_default.go |
| `/api/domain/{id}/files` | GET | handleDomainFilesList | application/json | ❌ | http.Error | json, file | routes_domain_files.go |
| `/api/domain/{id}/file/{filename}` | GET | handleDomainFileGet | dynamic MIME | ❌ | http.Error | file, json | routes_domain_files.go |
| `/api/domain/{id}/file/{filename}` | PUT | handleDomainFilePut | text/plain | ❌ | http.Error | json | routes_domain_files.go |
| `/api/domain/{id}/file/{filename}` | DELETE | handleDomainFileDelete | text/plain | ❌ | http.Error | json | routes_domain_files.go |
| `/api/domain/{id}/file/{filename}` | POST | handleDomainFileCreate | text/plain | ❌ | http.Error | json | routes_domain_files.go |
| `/api/domain/{id}/announce` | GET | handleDomainAnnounce | text/plain | ❌ | http.Error | json, text | routes.go |
| `/api/domain/{id}/css` | GET | handleDomainCSS | text/css | ❌ | http.Error | file, json | routes_domain_css.go |
| `/api/domain/{id}/css` | POST | handleUpdateDomainCSS | application/json | ❌ | http.Error | json | domain.go |
| `/api/domain/{id}/file/rename` | POST | handleDomainFileRename | text/plain | ❌ | http.Error | json | routes_domain_files.go |
| `/api/authority/console/evaluate` | POST | handleAuthorityEvaluate | application/json | ❌ | manual JSON | json | routes_authority.go |
| `/api/authority/console/schema` | GET | handleAuthoritySchema | application/json | ❌ | manual JSON | json, file | routes_authority.go |
| `/api/files/search` | GET | handleFileSearch | not implemented | ❌ | http.Error | json | files.go |
| `/api/files/export` | GET | handleFileExport | not implemented | ❌ | http.Error | json, file | files.go |
| `/api/files/` | GET | handleFileByID | not implemented | ❌ | http.Error | file, json | files.go |
| `/api/files` | GET | handleListFiles | not implemented | ❌ | http.Error | json | files.go |
| `/api/policy/list` | GET | handleListPolicies | application/json | ❌ | http.Error | json, file | routes_policy_files.go |
| `/api/policy/{name}` | GET | handleGetPolicy | text/plain | ❌ | http.Error | file, json | routes_policy_files.go |
| `/api/policy/{name}` | PUT | handlePutPolicy | text/plain | ❌ | http.Error | json | routes_policy_files.go |
| `/api/policy/{name}` | DELETE | handleDeletePolicy | text/plain | ❌ | http.Error | json | routes_policy_files.go |
| `/api/domain/{id}/policy` | GET | handleGetDomainPolicy | application/json | ❌ | http.Error | json, file | domain_policy.go |
| `/api/domain/{id}/policy` | POST | handleSetDomainPolicy | application/json | ❌ | http.Error | json | domain_policy.go |
| `/api/atlas/receipts` | GET | handleAtlasReceipts | not implemented | ❌ | http.Error | json | atlas/register.go |
| `/api/overlay/` | GET | handleOverlay | not implemented | ❌ | http.Error | json | atlas/register.go |
| `/api/version` | GET | anonymous | not implemented | ❌ | http.Error | json, text | routes_version.go |
| `/api/receipts/latest` | GET | handleGetLatestReceipt | not implemented | ❌ | http.Error | json | routes_receipts.go |
| `/api/identities` | GET | HandleIdentities | not implemented | ✅ | JSON helper | json | identities/register.go |
| `/api/net/peers` | GET | anonymous | not implemented | ❌ | http.Error | json | routes_net.go |
| `/api/schema/active` | GET | handleGetActiveSchema | not implemented | ❌ | http.Error | json | routes_schema.go |
| `/api/schema/list` | GET | handleListSchemas | not implemented | ❌ | http.Error | json | routes_schema.go |
| `/api/flow/rule` | GET | handleGetFlowRule | application/json | ✅ | JSON helper | json | flow.go |
| `/api/flow/rule/update` | POST | handleUpdateFlowRule | not implemented | ❌ | http.Error | json | flow.go |

## Key Findings

### Format Support Patterns
- **JSON-only endpoints**: 25+ endpoints returning application/json
- **File-serving endpoints**: 8 endpoints serving files with dynamic MIME types
- **Text-only endpoints**: 10+ endpoints returning text/plain or specific text types
- **Not implemented**: 12+ endpoints returning placeholder responses

### Error Handling Analysis
- **http.Error usage**: 95% of endpoints use non-JSON error responses
- **JSON helper usage**: Only 3 endpoints use standardized JSON() helper
- **Mixed error patterns**: Different error message formats across endpoints

### File Handler Patterns

#### Dynamic MIME Type Handlers
- **Domain Assets** (`/api/domain/{id}/file/{filename}`): CSS, JS, binary files
- **Domain CSS** (`/api/domain/{id}/css`): text/css with fallback
- **Policy Files** (`/api/policy/{name}`): text/plain for .rego files

#### Fixed Format Handlers
- **Domain Lists**: JSON arrays of domain objects
- **Status endpoints**: JSON status objects
- **API responses**: JSON success/error objects

## Content Negotiation Assessment

### Current State
- **No explicit negotiation**: All endpoints serve single format per route
- **No format suffixes**: No existing /{format} pattern usage
- **Route-based format**: Format determined by endpoint purpose, not client request

### Opportunities
- Add `/json` suffix for data endpoints
- Add `/file` suffix for raw content endpoints
- Add `/text` suffix for plain text responses
- Implement format-based error handling

## Refactor Complexity Analysis

### High-Impact Refactor Targets

#### 1. Error Handling Standardization (40+ locations)
**Current**: `http.Error(w, "message", statusCode)`
**Target**: `JSONError(w, statusCode, "message")` with format support
**Complexity**: High - affects every endpoint

#### 2. JSON Response Standardization (25+ endpoints)
**Current**: Manual `json.NewEncoder(w).Encode(data)`
**Target**: `JSON(w, statusCode, data)` with format switching
**Complexity**: Medium - straightforward refactor

#### 3. File Handler Format Support (8 endpoints)
**Current**: Fixed MIME types per file extension
**Target**: Format suffix support (json metadata vs raw file)
**Complexity**: Medium - requires format detection logic

### Low-Impact Refactor Targets

#### 1. Text Endpoint Unification (10+ endpoints)
**Current**: Various text response patterns
**Target**: Unified text/json format switching
**Complexity**: Low - simple format branching

#### 2. Not Implemented Handlers (12+ endpoints)
**Current**: Placeholder implementations
**Target**: Proper format-aware implementations
**Complexity**: Variable - depends on intended functionality

## Recommended Implementation Strategy

### Phase 1: Foundation (JSON Helper Standardization)
1. **Extend JSON helpers**: Add `JSONError()` and `JSONUnsupportedFormat()` functions
2. **Replace http.Error**: Systematically replace all `http.Error` calls
3. **Standardize JSON responses**: Replace manual json.NewEncoder usage

### Phase 2: Format Switching Infrastructure
1. **Add format detection**: Parse /{format} suffix from request paths
2. **Implement format switching**: Add conditional logic per handler type
3. **Add CBOR/encrypted support**: Implement non-JSON format handlers

### Phase 3: Handler Migration
1. **Data endpoints**: Add /json and /file suffix support
2. **File endpoints**: Add /json metadata vs /file content distinction
3. **Text endpoints**: Add /json structure vs /text raw content options

### Format Suffix Recommendations

| Endpoint Type | Suggested Formats | Primary Use Case |
|---------------|------------------|------------------|
| Domain data | /json, /file | API consumption vs download |
| Status/health | /json, /text | Structured data vs monitoring |
| File content | /file, /json | Raw content vs metadata |
| Lists/arrays | /json, /file | API response vs export |
| Policy content | /file, /json | Raw policy vs metadata |

## Implementation Notes

### JSON Helper Extensions Needed
```go
// Add to respond.go
func JSONError(w http.ResponseWriter, status int, message string) {
    JSON(w, status, map[string]any{
        "error": message,
        "status": status,
    })
}

func JSONUnsupportedFormat(w http.ResponseWriter, format string) {
    JSONError(w, http.StatusNotAcceptable,
        fmt.Sprintf("unsupported format: %s", format))
}
```

### Format Detection Pattern
```go
// Extract format from path suffix
func detectFormat(path string) string {
    if strings.HasSuffix(path, "/json") { return "json" }
    if strings.HasSuffix(path, "/file") { return "file" }
    if strings.HasSuffix(path, "/text") { return "text" }
    if strings.HasSuffix(path, "/cbor") { return "cbor" }
    if strings.HasSuffix(path, "/encrypted") { return "encrypted" }
    return "json" // default
}
```

### Handler Skeleton Template
```go
func (s *Server) handleExample(w http.ResponseWriter, r *http.Request) {
    format := detectFormat(r.URL.Path)

    // Business logic here
    data, err := s.getData()
    if err != nil {
        switch format {
        case "json":
            JSONError(w, http.StatusInternalServerError, err.Error())
        default:
            JSONUnsupportedFormat(w, format)
        }
        return
    }

    // Format-specific response
    switch format {
    case "json":
        JSON(w, http.StatusOK, data)
    case "file":
        serveAsFile(w, data)
    case "text":
        serveAsText(w, data)
    default:
        JSONUnsupportedFormat(w, format)
    }
}
```

## Conclusion

The DIS-Core API has a solid foundation with consistent route registration patterns but lacks format standardization. The primary refactor challenges are:

1. **Error handling inconsistency**: 95% of endpoints use non-JSON errors
2. **Manual JSON encoding**: Most endpoints use direct json.NewEncoder instead of helpers
3. **No format negotiation**: Single format per endpoint limits client flexibility

The recommended phased approach will systematically address these issues while maintaining API compatibility and enabling the unified /{format} suffix pattern across all endpoints.

---
*Investigation completed: November 10, 2025*
