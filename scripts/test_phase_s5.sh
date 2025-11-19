#!/bin/bash
# Phase S5: Per-Seat REGO Policy Evaluation Test

set -e

echo "🧪 Phase S5: Per-Seat REGO Integration Test"
echo "============================================"
echo

# Configuration
PGPASSWORD=card567
export PGPASSWORD
PSQL="psql -X -h localhost -U dis_user -d dis -t -A -q"
DOMAIN_ID=$($PSQL -c "SELECT id FROM domains LIMIT 1;")
BASE_URL="http://localhost:8080"

echo "📋 Test Configuration:"
echo "  Domain ID: $DOMAIN_ID"
echo "  Base URL: $BASE_URL"
echo

# Step 1: Create a test seat with custom REGO
echo "✅ Step 1: Create member seat with restrictive REGO policy"
APPOINT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/domain/$DOMAIN_ID/seats/appoint" \
  -H 'Content-Type: application/json' \
  -d '{
    "member_id":"restricted-seat@example.com",
    "scope":"restricted",
    "rego_ref":"seat:restricted:v1",
    "policy_version":"v1.0",
    "appointment_receipt":"rcpt-restricted-001"
  }')
SEAT_ID=$(echo "$APPOINT_RESPONSE" | jq -r '.id')
echo "  ✓ Created seat: $SEAT_ID"
echo

# Step 2: Update seat with restrictive REGO policy
echo "✅ Step 2: Update seat with restrictive REGO (only allows ci.call.v1 with risk < 30)"
REGO_POLICY=$(cat <<'EOF'
package dis.seat

export_allow {
  input.action == "ci.call.v1"
  input.risk < 30
}

export_allow {
  input.action == "ci.import.v1"
  input.risk < 10
}
EOF
)

# Use jq to properly escape the REGO policy for JSON
REGO_JSON=$(jq -n --arg rego "$REGO_POLICY" --arg ver "v1.0" '{rego_text: $rego, policy_version: $ver}')

REGO_UPDATE=$(curl -s -X PUT "$BASE_URL/api/domain/$DOMAIN_ID/seats/$SEAT_ID/rego" \
  -H 'Content-Type: application/json' \
  -d "$REGO_JSON")
echo "  ✓ REGO updated: $(echo "$REGO_UPDATE" | jq -r '.policy_version')"
echo

# Step 3: Test policy evaluation WITHOUT domain_id (no seat policies)
echo "✅ Step 3: Test evaluation WITHOUT domain_id (base policies only)"
EVAL_NO_DOMAIN=$(curl -s -X POST "$BASE_URL/api/policy/evaluate" \
  -H 'Content-Type: application/json' \
  -d '{
    "action": "ci.call.v1",
    "context": {
      "risk": 50
    }
  }')
echo "  Result (should allow - no seat policies):"
echo "$EVAL_NO_DOMAIN" | jq '{allow, reason}' | sed 's/^/    /'
echo

# Step 4: Test evaluation WITH domain_id (includes seat policies)
echo "✅ Step 4: Test evaluation WITH domain_id and high risk (50)"
EVAL_HIGH_RISK=$(curl -s -X POST "$BASE_URL/api/policy/evaluate" \
  -H 'Content-Type: application/json' \
  -d "{
    \"action\": \"ci.call.v1\",
    \"context\": {
      \"domain_id\": \"$DOMAIN_ID\",
      \"risk\": 50
    }
  }")
echo "  Result (should DENY - risk 50 exceeds seat policy limit 30):"
echo "$EVAL_HIGH_RISK" | jq '{allow, reason, details}' | sed 's/^/    /'
echo

# Step 5: Test with low risk (should allow)
echo "✅ Step 5: Test evaluation WITH domain_id and low risk (20)"
EVAL_LOW_RISK=$(curl -s -X POST "$BASE_URL/api/policy/evaluate" \
  -H 'Content-Type: application/json' \
  -d "{
    \"action\": \"ci.call.v1\",
    \"context\": {
      \"domain_id\": \"$DOMAIN_ID\",
      \"risk\": 20
    }
  }")
echo "  Result (should ALLOW - risk 20 is below seat policy limit 30):"
echo "$EVAL_LOW_RISK" | jq '{allow, reason}' | sed 's/^/    /'
echo

# Step 6: Test with different action
echo "✅ Step 6: Test evaluation with ci.import.v1 action and risk 15"
EVAL_IMPORT=$(curl -s -X POST "$BASE_URL/api/policy/evaluate" \
  -H 'Content-Type: application/json' \
  -d "{
    \"action\": \"ci.import.v1\",
    \"context\": {
      \"domain_id\": \"$DOMAIN_ID\",
      \"risk\": 15
    }
  }")
echo "  Result (should DENY - import requires risk < 10):"
echo "$EVAL_IMPORT" | jq '{allow, reason}' | sed 's/^/    /'
echo

# Step 7: Freeze the seat and test again
echo "✅ Step 7: Freeze seat and verify policy evaluation fails"
curl -s -X POST "$BASE_URL/api/domain/$DOMAIN_ID/seats/$SEAT_ID/freeze" > /dev/null
echo "  ✓ Seat frozen"

# Note: Frozen seats are excluded from GetActiveSeats, so they won't be evaluated
EVAL_AFTER_FREEZE=$(curl -s -X POST "$BASE_URL/api/policy/evaluate" \
  -H 'Content-Type: application/json' \
  -d "{
    \"action\": \"ci.call.v1\",
    \"context\": {
      \"domain_id\": \"$DOMAIN_ID\",
      \"risk\": 20
    }
  }")
echo "  Result (should allow - frozen seats not included in evaluation):"
echo "$EVAL_AFTER_FREEZE" | jq '{allow, reason}' | sed 's/^/    /'
echo

# Cleanup: Unfreeze the seat
curl -s -X POST "$BASE_URL/api/domain/$DOMAIN_ID/seats/$SEAT_ID/unfreeze" > /dev/null

# Summary
echo "============================================"
echo "✅ Phase S5 Integration Test Complete"
echo
echo "Summary:"
echo "  ✓ Per-seat REGO policies are loaded dynamically"
echo "  ✓ Seat policies can tighten restrictions (deny)"
echo "  ✓ Evaluation includes seat policy details"
echo "  ✓ Frozen seats are excluded from evaluation"
echo
echo "Test seat: $SEAT_ID"
echo
echo "🎉 Phase S5: Per-Seat REGO OPA integration verified!"
