#!/bin/bash
# Test script for Phase 10L: Guided Policy Correction

API_BASE="http://localhost:8080"
ROOT_DOMAIN="00000000-0000-0000-0000-000000000000"
CHILD_DOMAIN="77de077e-3099-4393-b99b-5724451003d7"

echo "=== Phase 10L Test Suite: Guided Policy Correction ==="
echo ""

echo "Test 1: Valid Rego v1 syntax - should pass"
echo "----------------------------------------"
curl -s -X POST "$API_BASE/api/policy/validate/$ROOT_DOMAIN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "package dis.policy\n\ndefault allow := false\n\nallow if {\n  input.action == \"read\"\n}"
  }' | jq -r '. | "Success: \(.success)\nDetails: \(.details)\nHints: \(.hints // [])"'
echo ""

echo "Test 2: Missing := operator (should suggest fix)"
echo "------------------------------------------------"
curl -s -X POST "$API_BASE/api/policy/validate/$ROOT_DOMAIN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "package dis.policy\n\ndefault allow = false\n\nallow if {\n  input.action == \"read\"\n}"
  }' | jq '.'
echo ""

echo "Test 3: Missing if keyword (should suggest fix)"
echo "-----------------------------------------------"
curl -s -X POST "$API_BASE/api/policy/validate/$ROOT_DOMAIN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "package dis.policy\n\ndefault allow := false\n\nallow {\n  input.action == \"read\"\n}"
  }' | jq '.'
echo ""

echo "Test 4: Undefined import (should detect missing parent)"
echo "-------------------------------------------------------"
curl -s -X POST "$API_BASE/api/policy/validate/$CHILD_DOMAIN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "package dis.policy\n\nimport data.dis.null\n\nallow if {\n  data.dis.null.allow\n}"
  }' | jq '.'
echo ""

echo "Test 5: Multiple errors (should return all hints)"
echo "-------------------------------------------------"
curl -s -X POST "$API_BASE/api/policy/validate/$ROOT_DOMAIN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "package dis.policy\n\ndefault allow = false\n\nallow {\n  input.action == \"read\"\n}\n\ndenial {\n  bad syntax here\n}"
  }' | jq '.'
echo ""

echo "Test 6: Valid child policy (inherits from valid parent)"
echo "-------------------------------------------------------"
curl -s -X POST "$API_BASE/api/policy/validate/$CHILD_DOMAIN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "package dis.policy\n\n# Child-specific rule\nallow if {\n  input.action == \"write\"\n  input.actor == \"admin\"\n}"
  }' | jq -r '. | "Success: \(.success)\nDetails: \(.details)\nHints: \(.hints // [])"'
echo ""

echo "=== Test Suite Complete ==="
