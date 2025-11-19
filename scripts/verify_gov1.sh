#!/bin/bash
# GOV-1 Verification Script
# Tests: Identity triad, domain creation, authority flow, REGO stack

set -e

echo "========================================"
echo "🏛️  GOV-1 Verification Script"
echo "========================================"
echo ""

BASE_URL="${DIS_API_URL:-http://localhost:8080}"
PASS=0
FAIL=0

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
pass_test() {
    echo -e "${GREEN}✅ $1${NC}"
    ((PASS++))
}

fail_test() {
    echo -e "${RED}❌ $1${NC}"
    ((FAIL++))
}

info_test() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Test 1: Check database migration (identity_seats table exists)
echo "Test 1: Database schema verification..."
if command -v psql &> /dev/null; then
    TABLE_EXISTS=$(psql "$DIS_DB_DSN" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'identity_seats');" 2>/dev/null || echo "f")
    if [ "$TABLE_EXISTS" = " t" ]; then
        pass_test "identity_seats table exists"
    else
        fail_test "identity_seats table not found"
    fi
else
    info_test "psql not available, skipping database check"
fi

# Test 2: Verify REGO policy files exist
echo ""
echo "Test 2: REGO policy stack verification..."
if [ -f "policy/terra.rego" ]; then
    pass_test "terra.rego exists"
else
    fail_test "terra.rego not found"
fi

if [ -f "policy/numen.rego" ]; then
    pass_test "numen.rego exists"
else
    fail_test "numen.rego not found"
fi

if [ -f "policy/lima.rego" ]; then
    pass_test "lima.rego exists"
else
    fail_test "lima.rego not found"
fi

# Test 3: Check Go code compilation
echo ""
echo "Test 3: Go code compilation..."
if go build -o /tmp/dis-core-gov1-test ./cmd/dis-core/ 2>/dev/null; then
    pass_test "Code compiles successfully"
    rm -f /tmp/dis-core-gov1-test
else
    fail_test "Compilation failed"
fi

# Test 4: Verify internal packages exist
echo ""
echo "Test 4: Package structure verification..."
if [ -d "internal/identity" ]; then
    pass_test "internal/identity package exists"
else
    fail_test "internal/identity package not found"
fi

if [ -d "internal/authority" ]; then
    pass_test "internal/authority package exists"
else
    fail_test "internal/authority package not found"
fi

# Test 5: Check bootstrap file
echo ""
echo "Test 5: Bootstrap code verification..."
if [ -f "cmd/dis-core/bootstrap/identities.go" ]; then
    pass_test "identities.go bootstrap exists"
else
    fail_test "identities.go bootstrap not found"
fi

# Test 6: Verify migration file
echo ""
echo "Test 6: Migration file verification..."
if [ -f "db/migrations/20251112_add_identity_seats_table.sql" ]; then
    pass_test "Identity seats migration exists"
else
    fail_test "Identity seats migration not found"
fi

# Test 7: Check for terra/numen/lima in REGO files
echo ""
echo "Test 7: REGO content verification..."
if grep -q "package dis.terra" policy/terra.rego 2>/dev/null; then
    pass_test "terra.rego has correct package"
else
    fail_test "terra.rego package incorrect"
fi

if grep -q "package dis.numen" policy/numen.rego 2>/dev/null; then
    pass_test "numen.rego has correct package"
else
    fail_test "numen.rego package incorrect"
fi

if grep -q "package dis.lima" policy/lima.rego 2>/dev/null; then
    pass_test "lima.rego has correct package"
else
    fail_test "lima.rego package incorrect"
fi

# Test 8: Verify REGO imports (numen imports terra, lima imports both)
echo ""
echo "Test 8: REGO layer dependencies..."
if grep -q "import data.dis.terra" policy/numen.rego 2>/dev/null; then
    pass_test "numen.rego imports terra"
else
    fail_test "numen.rego missing terra import"
fi

if grep -q "import data.dis.terra" policy/lima.rego 2>/dev/null && \
   grep -q "import data.dis.numen" policy/lima.rego 2>/dev/null; then
    pass_test "lima.rego imports terra and numen"
else
    fail_test "lima.rego missing imports"
fi

# Test 9: Check authority flow engine
echo ""
echo "Test 9: Authority flow engine verification..."
if grep -q "func EvaluateAuthority" internal/authority/flow_engine.go 2>/dev/null; then
    pass_test "EvaluateAuthority function exists"
else
    fail_test "EvaluateAuthority function not found"
fi

# Test 10: Verify triad repository methods
echo ""
echo "Test 10: Triad repository verification..."
if grep -q "func.*InitializeTriad" internal/identity/triad_repo.go 2>/dev/null; then
    pass_test "InitializeTriad method exists"
else
    fail_test "InitializeTriad method not found"
fi

if grep -q "func.*GetIdentityTriad" internal/identity/triad_repo.go 2>/dev/null; then
    pass_test "GetIdentityTriad method exists"
else
    fail_test "GetIdentityTriad method not found"
fi

# Test 11: Check for seat state constants
echo ""
echo "Test 11: Seat state constants..."
if grep -q "SeatStateEmpty\|SeatStateAssigned\|SeatStateOccupied\|SeatStateFrozen" internal/identity/triad_model.go 2>/dev/null; then
    pass_test "Seat state constants defined"
else
    fail_test "Seat state constants missing"
fi

# Test 12: Check for seat type constants
if grep -q "SeatTypeTerra\|SeatTypeNumen\|SeatTypeLima" internal/identity/triad_model.go 2>/dev/null; then
    pass_test "Seat type constants defined"
else
    fail_test "Seat type constants missing"
fi

# Summary
echo ""
echo "========================================"
echo "📊 Results: $PASS passed, $FAIL failed"
echo "========================================"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✅ All GOV-1 foundation tests passed!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Run database migration: psql \$DIS_DB_DSN < db/migrations/20251112_add_identity_seats_table.sql"
    echo "  2. Start server: go run cmd/dis-core/main.go"
    echo "  3. Bootstrap will auto-create identity triads"
    echo "  4. Test with: curl http://localhost:8080/api/identity/{id}"
    echo ""
    exit 0
else
    echo -e "${RED}❌ Some GOV-1 tests failed${NC}"
    echo ""
    echo "Review the failures above and fix before proceeding."
    echo ""
    exit 1
fi
