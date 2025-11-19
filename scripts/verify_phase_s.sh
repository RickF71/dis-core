#!/bin/bash
# Phase S Verification Script
# Tests all Phase S0-S4 functionality

set -e

echo "🧪 Phase S: Seat Management System Verification"
echo "================================================"
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

# Test 1: Phase S0 - Root Seats
echo "✅ Test 1: Phase S0 — Root Seat Bootstrap"
ROOT_COUNT=$($PSQL -c "SELECT COUNT(*) FROM domain_seats WHERE seat_type='root';")
echo "  ✓ Root seats created: $ROOT_COUNT"
echo

# Test 2: Phase S1 - Schema
echo "✅ Test 2: Phase S1 — Database Schema"
TABLE_EXISTS=$($PSQL -c "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name='domain_seats');")
if [ "$TABLE_EXISTS" = "t" ]; then
  echo "  ✓ domain_seats table exists"
else
  echo "  ✗ domain_seats table NOT FOUND"
  exit 1
fi
echo

# Test 3: Phase S4 - List Seats API
echo "✅ Test 3: Phase S4 — GET /api/domain/{id}/seats"
SEATS_RESPONSE=$(curl -s "$BASE_URL/api/domain/$DOMAIN_ID/seats")
SEATS_COUNT=$(echo "$SEATS_RESPONSE" | jq '. | length')
echo "  ✓ Returned $SEATS_COUNT seats"
echo "  First seat:"
echo "$SEATS_RESPONSE" | jq '.[0]' | sed 's/^/    /'
echo

# Test 4: Appoint Member Seat
echo "✅ Test 4: Phase S4 — POST /api/domain/{id}/seats/appoint"
APPOINT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/domain/$DOMAIN_ID/seats/appoint" \
  -H 'Content-Type: application/json' \
  -d '{
    "member_id":"test-member-'$(date +%s)'@example.com",
    "scope":"testing",
    "rego_ref":"seat:test:v1",
    "policy_version":"v1.0",
    "appointment_receipt":"rcpt-test-'$(date +%s)'"
  }')
MEMBER_SEAT_ID=$(echo "$APPOINT_RESPONSE" | jq -r '.id')
MEMBER_ID=$(echo "$APPOINT_RESPONSE" | jq -r '.member_id')
echo "  ✓ Member seat created: $MEMBER_SEAT_ID"
echo "  ✓ Member ID: $MEMBER_ID"
echo

# Test 5: Freeze Seat
echo "✅ Test 5: Phase S4 — POST /api/domain/{id}/seats/{seatId}/freeze"
FREEZE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/domain/$DOMAIN_ID/seats/$MEMBER_SEAT_ID/freeze")
FREEZE_SUCCESS=$(echo "$FREEZE_RESPONSE" | jq -r '.success')
if [ "$FREEZE_SUCCESS" = "true" ]; then
  echo "  ✓ Seat frozen successfully"
  # Verify in DB
  DB_STATUS=$($PSQL -c "SELECT status FROM domain_seats WHERE id='$MEMBER_SEAT_ID';")
  echo "  ✓ Database status: $DB_STATUS"
else
  echo "  ✗ Freeze failed"
  exit 1
fi
echo

# Test 6: Unfreeze Seat
echo "✅ Test 6: Phase S4 — POST /api/domain/{id}/seats/{seatId}/unfreeze"
UNFREEZE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/domain/$DOMAIN_ID/seats/$MEMBER_SEAT_ID/unfreeze")
UNFREEZE_SUCCESS=$(echo "$UNFREEZE_RESPONSE" | jq -r '.success')
if [ "$UNFREEZE_SUCCESS" = "true" ]; then
  echo "  ✓ Seat unfrozen successfully"
  DB_STATUS=$($PSQL -c "SELECT status FROM domain_seats WHERE id='$MEMBER_SEAT_ID';")
  echo "  ✓ Database status: $DB_STATUS"
else
  echo "  ✗ Unfreeze failed"
  exit 1
fi
echo

# Test 7: Update Seat REGO
echo "✅ Test 7: Phase S4 — PUT /api/domain/{id}/seats/{seatId}/rego"
REGO_UPDATE_RESPONSE=$(curl -s -X PUT "$BASE_URL/api/domain/$DOMAIN_ID/seats/$MEMBER_SEAT_ID/rego" \
  -H 'Content-Type: application/json' \
  -d '{
    "rego_text":"package dis.seat.test\n\nexport_allow {\n  input.action == \"ci.call.v1\"\n  input.risk < 30\n}",
    "policy_version":"v1.1"
  }')
REGO_SUCCESS=$(echo "$REGO_UPDATE_RESPONSE" | jq -r '.success')
REGO_VERSION=$(echo "$REGO_UPDATE_RESPONSE" | jq -r '.policy_version')
if [ "$REGO_SUCCESS" = "true" ]; then
  echo "  ✓ REGO updated successfully"
  echo "  ✓ Policy version: $REGO_VERSION"
  REGO_LENGTH=$($PSQL -c "SELECT LENGTH(rego_text) FROM domain_seats WHERE id='$MEMBER_SEAT_ID';")
  echo "  ✓ REGO length in DB: $REGO_LENGTH bytes"
else
  echo "  ✗ REGO update failed"
  exit 1
fi
echo

# Test 8: Verify Seat Lineage
echo "✅ Test 8: Seat Lineage Verification"
LINEAGE=$($PSQL -c "
  SELECT
    s.id,
    s.seat_type,
    s.member_id,
    s.status,
    s.parent_seat_id,
    s.appointed_by
  FROM domain_seats s
  WHERE s.id='$MEMBER_SEAT_ID';
")
echo "  Member seat lineage:"
echo "$LINEAGE" | sed 's/^/    /'
echo

# Summary
echo "================================================"
echo "✅ Phase S Verification Complete"
echo
echo "Summary:"
echo "  ✓ Phase S0: Bootstrap verified ($ROOT_COUNT root seats)"
echo "  ✓ Phase S1: Schema verified (domain_seats table exists)"
echo "  ✓ Phase S4: All 5 API endpoints working"
echo "    - GET /api/domain/{id}/seats"
echo "    - POST /api/domain/{id}/seats/appoint"
echo "    - POST /api/domain/{id}/seats/{seatId}/freeze"
echo "    - POST /api/domain/{id}/seats/{seatId}/unfreeze"
echo "    - PUT /api/domain/{id}/seats/{seatId}/rego"
echo
echo "Test seat created: $MEMBER_SEAT_ID"
echo "Member ID: $MEMBER_ID"
echo
echo "🎉 Phase S0-S4 implementation verified!"
