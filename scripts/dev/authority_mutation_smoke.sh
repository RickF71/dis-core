#!/usr/bin/env bash
# GOV-3: Smoke test script for seat transition write APIs
set -euo pipefail

BASE="${1:-http://localhost:8080}"
IDENTITY_ID="${IDENTITY_ID:?Error: set IDENTITY_ID environment variable}"
ACTOR_ID="${ACTOR_ID:-$IDENTITY_ID}"
DOMAIN_ID="${DOMAIN_ID:-terra}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== GOV-3 Seat Mutation Smoke Tests ===${NC}"
echo "Base URL: $BASE"
echo "Identity: $IDENTITY_ID"
echo "Actor: $ACTOR_ID"
echo "Domain: $DOMAIN_ID"
echo ""

# Test 1: Transition EMPTY → ASSIGNED
echo -e "${YELLOW}[1/4] Test: EMPTY → ASSIGNED${NC}"
RESULT=$(curl -sS -X POST "$BASE/api/authority/seat/transition" \
  -H 'Content-Type: application/json' \
  -d "{
    \"domain_id\": \"$DOMAIN_ID\",
    \"identity_id\": \"$IDENTITY_ID\",
    \"layer\": \"lima\",
    \"from\": \"EMPTY\",
    \"to\": \"ASSIGNED\",
    \"actor_id\": \"$ACTOR_ID\",
    \"reason\": \"smoke test: bootstrap assignment\",
    \"context\": {
      \"permissions\": [\"seat.transition\"]
    }
  }")

echo "$RESULT" | jq .

if echo "$RESULT" | jq -e '.ok == true' > /dev/null 2>&1; then
  echo -e "${GREEN}✅ PASSED: Transition to ASSIGNED succeeded${NC}"
else
  echo -e "${RED}❌ FAILED: Transition to ASSIGNED failed${NC}"
  echo "$RESULT" | jq .
fi
echo ""

# Wait a moment for state to settle
sleep 1

# Test 2: Transition ASSIGNED → OCCUPIED
echo -e "${YELLOW}[2/4] Test: ASSIGNED → OCCUPIED${NC}"
RESULT=$(curl -sS -X POST "$BASE/api/authority/seat/transition" \
  -H 'Content-Type: application/json' \
  -d "{
    \"domain_id\": \"$DOMAIN_ID\",
    \"identity_id\": \"$IDENTITY_ID\",
    \"layer\": \"lima\",
    \"from\": \"ASSIGNED\",
    \"to\": \"OCCUPIED\",
    \"actor_id\": \"$ACTOR_ID\",
    \"reason\": \"smoke test: enable consent authority\",
    \"context\": {
      \"permissions\": [\"seat.transition\"]
    }
  }")

echo "$RESULT" | jq .

if echo "$RESULT" | jq -e '.ok == true' > /dev/null 2>&1; then
  echo -e "${GREEN}✅ PASSED: Transition to OCCUPIED succeeded${NC}"
  DECISION_ID=$(echo "$RESULT" | jq -r '.decision_id // empty')
  RECEIPT_ID=$(echo "$RESULT" | jq -r '.receipt_id // empty')
  echo "  Decision ID: $DECISION_ID"
  echo "  Receipt ID: $RECEIPT_ID"
else
  echo -e "${RED}❌ FAILED: Transition to OCCUPIED failed${NC}"
  echo "$RESULT" | jq .
fi
echo ""

sleep 1

# Test 3: Transition OCCUPIED → FROZEN
echo -e "${YELLOW}[3/4] Test: OCCUPIED → FROZEN${NC}"
RESULT=$(curl -sS -X POST "$BASE/api/authority/seat/transition" \
  -H 'Content-Type: application/json' \
  -d "{
    \"domain_id\": \"$DOMAIN_ID\",
    \"identity_id\": \"$IDENTITY_ID\",
    \"layer\": \"lima\",
    \"from\": \"OCCUPIED\",
    \"to\": \"FROZEN\",
    \"actor_id\": \"$ACTOR_ID\",
    \"reason\": \"smoke test: freeze seat\",
    \"context\": {
      \"permissions\": [\"seat.transition\"]
    }
  }")

echo "$RESULT" | jq .

if echo "$RESULT" | jq -e '.ok == true' > /dev/null 2>&1; then
  echo -e "${GREEN}✅ PASSED: Transition to FROZEN succeeded${NC}"
else
  echo -e "${RED}❌ FAILED: Transition to FROZEN failed${NC}"
  echo "$RESULT" | jq .
fi
echo ""

sleep 1

# Test 4: Thaw FROZEN → ASSIGNED (break-glass)
echo -e "${YELLOW}[4/4] Test: FROZEN → ASSIGNED (break-glass thaw)${NC}"
RESULT=$(curl -sS -X POST "$BASE/api/authority/seat/transition" \
  -H 'Content-Type: application/json' \
  -d "{
    \"domain_id\": \"$DOMAIN_ID\",
    \"identity_id\": \"$IDENTITY_ID\",
    \"layer\": \"lima\",
    \"from\": \"FROZEN\",
    \"to\": \"ASSIGNED\",
    \"actor_id\": \"$ACTOR_ID\",
    \"reason\": \"smoke test: thaw frozen seat\",
    \"context\": {
      \"permissions\": [\"seat.transition\"]
    }
  }")

echo "$RESULT" | jq .

if echo "$RESULT" | jq -e '.ok == true' > /dev/null 2>&1; then
  echo -e "${GREEN}✅ PASSED: Thaw to ASSIGNED succeeded${NC}"
else
  echo -e "${RED}❌ FAILED: Thaw to ASSIGNED failed${NC}"
  echo "$RESULT" | jq .
fi
echo ""

# Test 5: Batch transition (optional)
echo -e "${YELLOW}[5/5] Test: Batch transition (terra + numen)${NC}"
RESULT=$(curl -sS -X POST "$BASE/api/authority/seat/transition/batch" \
  -H 'Content-Type: application/json' \
  -d "{
    \"items\": [
      {
        \"domain_id\": \"$DOMAIN_ID\",
        \"identity_id\": \"$IDENTITY_ID\",
        \"layer\": \"terra\",
        \"from\": \"EMPTY\",
        \"to\": \"ASSIGNED\",
        \"actor_id\": \"$ACTOR_ID\",
        \"reason\": \"batch test: terra assignment\",
        \"context\": {\"permissions\": [\"seat.transition\"]}
      },
      {
        \"domain_id\": \"$DOMAIN_ID\",
        \"identity_id\": \"$IDENTITY_ID\",
        \"layer\": \"numen\",
        \"from\": \"EMPTY\",
        \"to\": \"ASSIGNED\",
        \"actor_id\": \"$ACTOR_ID\",
        \"reason\": \"batch test: numen assignment\",
        \"context\": {\"permissions\": [\"seat.transition\"]}
      }
    ]
  }")

echo "$RESULT" | jq .

SUCCESS_COUNT=$(echo "$RESULT" | jq '[.results[] | select(.ok == true)] | length')
TOTAL_COUNT=$(echo "$RESULT" | jq '.results | length')

if [ "$SUCCESS_COUNT" -eq "$TOTAL_COUNT" ]; then
  echo -e "${GREEN}✅ PASSED: Batch transition succeeded ($SUCCESS_COUNT/$TOTAL_COUNT)${NC}"
else
  echo -e "${YELLOW}⚠️  PARTIAL: Batch transition ($SUCCESS_COUNT/$TOTAL_COUNT succeeded)${NC}"
fi
echo ""

echo -e "${BLUE}=== Smoke Tests Complete ===${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Check decision and receipt IDs in your audit logs"
echo "2. Open Finagler Authority Console to see real-time updates"
echo "3. Monitor SSE stream: curl -N $BASE/api/authority/events"
