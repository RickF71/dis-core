#!/usr/bin/env bash
# GOV-2 Authority Console API Smoke Test
# Tests: Identity triad, flow preview, flow evaluation

set -euo pipefail

BASE_URL="${DIS_API_URL:-http://localhost:8080}"
IDENTITY_ID="${IDENTITY_ID:-test@example.com}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "========================================"
echo -e "${BLUE}🏛️  GOV-2 Authority Console API Tests${NC}"
echo "========================================"
echo ""
echo "Base URL: $BASE_URL"
echo "Identity: $IDENTITY_ID"
echo ""

# Test 1: Health check
echo -e "${YELLOW}Test 1: Health check...${NC}"
if curl -sS -f "$BASE_URL/api/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Server is responding${NC}"
else
    echo -e "${RED}❌ Server not responding${NC}"
    exit 1
fi
echo ""

# Test 2: Fetch identity triad
echo -e "${YELLOW}Test 2: Fetch identity triad...${NC}"
echo "GET $BASE_URL/api/authority/triad/$IDENTITY_ID"
TRIAD_RESPONSE=$(curl -sS "$BASE_URL/api/authority/triad/$IDENTITY_ID" 2>&1)
if echo "$TRIAD_RESPONSE" | jq . > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Triad endpoint responding${NC}"
    echo "$TRIAD_RESPONSE" | jq '.seats[] | {layer, state}'
else
    echo -e "${RED}❌ Triad fetch failed${NC}"
    echo "$TRIAD_RESPONSE"
fi
echo ""

# Test 3: Flow preview
echo -e "${YELLOW}Test 3: Flow preview (sample scenarios)...${NC}"
echo "GET $BASE_URL/api/authority/flow/preview"
PREVIEW_RESPONSE=$(curl -sS "$BASE_URL/api/authority/flow/preview" 2>&1)
if echo "$PREVIEW_RESPONSE" | jq . > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Flow preview endpoint responding${NC}"
    SCENARIO_COUNT=$(echo "$PREVIEW_RESPONSE" | jq '.scenarios | length')
    echo "   Found $SCENARIO_COUNT sample scenarios"
    echo "$PREVIEW_RESPONSE" | jq '.scenarios[] | {name, expected}'
else
    echo -e "${RED}❌ Flow preview failed${NC}"
    echo "$PREVIEW_RESPONSE"
fi
echo ""

# Test 4: Flow evaluation (upward reporting - should allow)
echo -e "${YELLOW}Test 4: Flow eval - upward reporting...${NC}"
EVAL_PAYLOAD=$(cat <<EOF
{
  "identity_id": "$IDENTITY_ID",
  "domain_id": "terra",
  "action": "report.status",
  "direction": "upward",
  "parent_approved": false
}
EOF
)

echo "POST $BASE_URL/api/authority/flow/eval"
EVAL_RESPONSE=$(curl -sS -X POST "$BASE_URL/api/authority/flow/eval" \
    -H 'Content-Type: application/json' \
    -d "$EVAL_PAYLOAD" 2>&1)

if echo "$EVAL_RESPONSE" | jq . > /dev/null 2>&1; then
    ALLOW=$(echo "$EVAL_RESPONSE" | jq -r '.allow')
    REASON=$(echo "$EVAL_RESPONSE" | jq -r '.reason')

    if [ "$ALLOW" == "true" ]; then
        echo -e "${GREEN}✅ Upward reporting ALLOWED (expected)${NC}"
    else
        echo -e "${YELLOW}⚠️  Upward reporting DENIED (check triad state)${NC}"
    fi
    echo "   Reason: $REASON"
    echo "$EVAL_RESPONSE" | jq '.triad_seats'
else
    echo -e "${RED}❌ Flow eval failed${NC}"
    echo "$EVAL_RESPONSE"
fi
echo ""

# Test 5: Flow evaluation (downward without OCCUPIED - should deny)
echo -e "${YELLOW}Test 5: Flow eval - downward governance...${NC}"
EVAL_DOWNWARD=$(cat <<EOF
{
  "identity_id": "$IDENTITY_ID",
  "domain_id": "terra",
  "action": "seat.appoint",
  "direction": "downward",
  "action_domain": "terra",
  "seat_domain": "terra",
  "parent_approved": true
}
EOF
)

echo "POST $BASE_URL/api/authority/flow/eval"
EVAL_DOWN_RESPONSE=$(curl -sS -X POST "$BASE_URL/api/authority/flow/eval" \
    -H 'Content-Type: application/json' \
    -d "$EVAL_DOWNWARD" 2>&1)

if echo "$EVAL_DOWN_RESPONSE" | jq . > /dev/null 2>&1; then
    ALLOW=$(echo "$EVAL_DOWN_RESPONSE" | jq -r '.allow')
    REASON=$(echo "$EVAL_DOWN_RESPONSE" | jq -r '.reason')

    echo "   Allow: $ALLOW"
    echo "   Reason: $REASON"

    # Check lima seat state
    LIMA_STATE=$(echo "$EVAL_DOWN_RESPONSE" | jq -r '.triad_seats[] | select(.layer=="lima") | .state')
    if [ "$LIMA_STATE" == "OCCUPIED" ]; then
        if [ "$ALLOW" == "true" ]; then
            echo -e "${GREEN}✅ Downward governance ALLOWED (lima OCCUPIED)${NC}"
        else
            echo -e "${YELLOW}⚠️  Downward governance DENIED despite OCCUPIED lima${NC}"
        fi
    else
        if [ "$ALLOW" == "false" ]; then
            echo -e "${GREEN}✅ Downward governance DENIED (lima not OCCUPIED)${NC}"
        else
            echo -e "${RED}❌ Downward governance ALLOWED without OCCUPIED lima${NC}"
        fi
    fi
else
    echo -e "${RED}❌ Flow eval (downward) failed${NC}"
    echo "$EVAL_DOWN_RESPONSE"
fi
echo ""

# Summary
echo "========================================"
echo -e "${BLUE}📊 GOV-2 API Smoke Test Complete${NC}"
echo "========================================"
echo ""
echo "Tested endpoints:"
echo "  • GET  /api/health"
echo "  • GET  /api/authority/triad/:identityId"
echo "  • GET  /api/authority/flow/preview"
echo "  • POST /api/authority/flow/eval"
echo ""
echo "Next steps:"
echo "  1. Check triad states in database:"
echo "     psql \$DIS_DB_DSN -c \"SELECT * FROM identity_seats WHERE identity_id='$IDENTITY_ID';\""
echo ""
echo "  2. Test with different identities:"
echo "     IDENTITY_ID=alice@example.com bash $0"
echo ""
echo "  3. Integrate Finagler UI components"
echo ""
