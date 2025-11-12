# Phase 10K: Finagler Rego Editor Integration

**Status**: ✅ Completed (including Root Domain Awareness extension)
**Date**: 2025-01-15

## Overview

Phase 10K implements a comprehensive Rego policy editor in Finagler with full OPA integration, enabling users to view inherited policies from parent domains and edit current domain policies with real-time validation.

**Extension**: Root domain awareness - the editor automatically detects parentless domains (root/null domains) and adapts the UI accordingly, hiding the inherited policy pane and showing a "Root Policy — no parent domain" badge.## Backend Implementation

### New API Endpoints

1. **GET `/api/domain/{id}/policy/inherited`**
   - Traverses domain lineage from current → parent → null
   - Concatenates all parent Rego policies with comment separators
   - Returns combined policy text for read-only viewing
   - Implements safety limit (20 levels) to prevent infinite loops

2. **GET `/api/domain/{id}/policy/current`**
   - Returns current domain's policy from `payload->policy->content`
   - Provides default Rego template if policy is empty:
     ```rego
     package dis.policy

     # Domain-specific policy rules
     default allow = false
     ```

3. **PATCH `/api/domain/{id}/policy`**
   - Updates domain policy content using `jsonb_set()`
   - Request body: `{"content": "package dis.policy\n..."}`
   - Updates both content and `updated_at` timestamp in JSONB
   - Returns success/error response

4. **POST `/api/policy/eval`**
   - Evaluates Rego policy against input data using OPA CLI
   - Request body:
     ```json
     {
       "policy": "package dis.policy\ndefault allow = false",
       "input": {"actor": "user@domain", "action": "read"},
       "query": "data.dis.policy.allow"
     }
     ```
   - Writes temporary files to `/tmp/policy.rego` and `/tmp/input.json`
   - Executes: `opa eval --data /tmp/policy.rego --input /tmp/input.json --format json {query}`
   - Returns: `{"result": true/false, "success": bool, "error": "...", "trace": []}`
   - Cleans up temp files after evaluation

### File: `internal/api/domain_policy.go`

**New Functions:**
- `GetDomainPolicyInherited(w, r)` - HTTP handler for inherited policies
- `GetDomainPolicyCurrent(w, r)` - HTTP handler for current policy
- `getDomainLineage(ctx, domainID)` - Recursive parent_id traversal
- `getDomainRegoPolicy(ctx, domainID)` - Queries `payload->'policy'->>'content'`

### File: `internal/api/policy_eval.go` (NEW)

**Structs:**
```go
type PolicyEvalRequest struct {
    Policy string         `json:"policy"`
    Input  map[string]any `json:"input"`
    Query  string         `json:"query"`
}

type PolicyEvalResponse struct {
    Result  any      `json:"result"`
    Success bool     `json:"success"`
    Error   string   `json:"error,omitempty"`
    Trace   []string `json:"trace,omitempty"`
}
```

**Functions:**
- `HandlePolicyEval(w, r)` - POST endpoint handler
- `evaluateRegoPolicy(policy, input, query)` - OPA CLI wrapper
- `UpdateDomainPolicy(w, r)` - PATCH endpoint handler

### Route Registration (`internal/api/routes.go`)

```go
// Phase 10K: Rego Policy Editor
r.Get("/api/domain/{id}/policy/inherited", s.GetDomainPolicyInherited)
r.Get("/api/domain/{id}/policy/current", s.GetDomainPolicyCurrent)
r.Patch("/api/domain/{id}/policy", s.UpdateDomainPolicy)
r.Post("/api/policy/eval", s.HandlePolicyEval)
```

## Frontend Implementation

### RegoEditor Component (`src/views/RegoEditor.jsx`)

**Features:**
- **Root Domain Awareness**: Automatically detects parentless domains (`parent_id === null`)
  - Shows amber badge: "Root Policy — no parent domain"
  - Hides inherited policy pane for root domains
  - Single editable Monaco panel for root domain policies
  - Skips `/policy/inherited` fetch when parent_id is null
- Dual Monaco Editor panels (for non-root domains):
  - **Top Panel**: Inherited policies (read-only, shows lineage from parents)
  - **Bottom Panel**: Current domain policy (editable)
- Action buttons:
  - **Validate**: Runs OPA evaluation with test input JSON
  - **Save Policy**: Updates domain policy in database, navigates to overview on success
- Test input panel: JSON editor for validation testing
- Validation result display: Shows OPA evaluation outcome with success/error state

**State Management:**
```javascript
const [inheritedPolicy, setInheritedPolicy] = useState("# Loading...");
const [currentPolicy, setCurrentPolicy] = useState("# Loading...");
const [validationResult, setValidationResult] = useState(null);
const [evalInput, setEvalInput] = useState('{"actor": "user@domain", ...}');
const isRootDomain = !domain?.parent_id;  // Root domain detection
```

**Monaco Configuration:**
- Custom Rego language definition registered
- Keywords: `package`, `import`, `default`, `else`, `with`, `some`, `in`, `not`
- Operators: `=`, `!=`, `==`, `<=`, `>=`, `:=`, `|`, `&`
- Syntax highlighting: comments (`#`), strings, numbers, brackets
- Auto-closing pairs for `{}`, `[]`, `()`, `""`
- Dark theme (`vs-dark`) for consistency

**Rego v1 Syntax:**
- OPA 1.9.0+ requires Rego v1 syntax
- Use `:=` for assignment instead of `=`
- Use `if` keyword before rule bodies: `allow if { ... }`
- Default policies use v1 syntax: `default allow := false`

### Navigation Integration

**Sidebar** (`src/components/Sidebar.jsx`):
- Added "Rego Editor" menu item
- View key: `"rego"`
- Position: After "Files", before "Identities"

**Routing** (`src/app/routes.jsx`):
- Imported `RegoEditor` component
- Added to `VIEW_MAP`: `rego: RegoEditor`

**DomainView Router** (`src/views/DomainView.jsx`):
- Added case: `case "rego": return <RegoEditor />;`

## Database Schema

Policy content is stored in the `domains.payload` JSONB field:

```json
{
  "css": {...},
  "meta": {...},
  "policy": {
    "content": "package dis.policy\n\ndefault allow = false",
    "updated_at": "2025-01-15T10:30:00Z"
  }
}
```

Query path: `domains.payload->'policy'->>'content'`

## OPA Integration

**Installation Requirement:**
- Requires OPA CLI installed on server
- Install: `curl -L -o opa https://openpolicyagent.org/downloads/latest/opa_linux_amd64`
- Make executable: `chmod +x opa && mv opa /usr/local/bin/`

**Execution Flow:**
1. Write policy to `/tmp/policy.rego`
2. Write input JSON to `/tmp/input.json`
3. Execute: `opa eval --data /tmp/policy.rego --input /tmp/input.json --format json data.dis.policy.allow`
4. Parse JSON output
5. Clean up temp files
6. Return result to frontend

**Error Handling:**
- Syntax errors in policy returned in `error` field
- Invalid input JSON caught before OPA execution
- Command execution failures logged and returned

## Testing Checklist

- [x] Backend endpoints implemented and routes registered
- [x] Frontend component created with dual Monaco panels
- [x] Monaco Rego syntax highlighting configured
- [x] Sidebar navigation link added
- [x] Routing integration complete
- [ ] Manual test: Load inherited policies from parent domains
- [ ] Manual test: Edit and save current domain policy
- [ ] Manual test: Validate policy with OPA using test input
- [ ] Manual test: Verify error handling for invalid Rego syntax
- [ ] Manual test: Confirm UI layout and responsiveness

## Usage Instructions

1. **Navigate to Rego Editor:**
   - Select a domain from the domain chooser
   - Click "Rego Editor" in the sidebar

2. **View Inherited Policies:**
   - Top panel shows combined policies from parent domains
   - Policies are concatenated with comment separators
   - Read-only to preserve lineage integrity

3. **Edit Current Policy:**
   - Bottom panel is editable
   - Syntax highlighting for Rego keywords
   - Auto-completion for brackets and quotes

4. **Validate Policy:**
   - Enter test input JSON (e.g., `{"actor": "user@domain", "action": "read"}`)
   - Click "Validate" button
   - View result (true/false/error) in validation panel

5. **Save Policy:**
   - Click "Save Policy" button
   - Toast notification confirms success
   - Policy stored in `domains.payload->policy->content`

## Example Policy

```rego
package dis.policy

# Import parent policies (inherited automatically)

# Default deny
default allow = false

# Allow read access for authenticated users
allow {
    input.action == "read"
    input.actor != ""
}

# Allow write access for admins
allow {
    input.action == "write"
    input.actor == "admin@example.com"
}
```

## Future Enhancements

- [ ] Policy templates dropdown (RBAC, ABAC, etc.)
- [ ] Policy diff view (compare current vs inherited)
- [ ] Policy version history
- [ ] IntelliSense for Rego rules
- [ ] Policy testing suite with predefined test cases
- [ ] Real-time validation as user types
- [ ] Policy impact analysis (show which resources affected)

## Related Documentation

- [OPA Documentation](https://www.openpolicyagent.org/docs/latest/)
- [Rego Language Reference](https://www.openpolicyagent.org/docs/latest/policy-language/)
- [Monaco Editor API](https://microsoft.github.io/monaco-editor/api/)
- [Phase 10J: Domain Payload Unification](./PHASE_10J.md)

---

**Implementation Notes:**
- All backend endpoints follow format-aware pattern (text responses)
- Frontend uses `react-hot-toast` for user notifications
- Monaco configuration runs before editor mount via `beforeMount` prop
- OPA evaluation is synchronous to prevent race conditions
- Temp files use `/tmp/` directory for security isolation
