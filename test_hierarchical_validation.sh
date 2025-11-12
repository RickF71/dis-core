#!/bin/bash
# Test script for hierarchical Rego validation

API_BASE="http://localhost:8080"
ROOT_DOMAIN="00000000-0000-0000-0000-000000000000"
CHILD_DOMAIN="77de077e-3099-4393-b99b-5724451003d7"

echo "=== Test 1: Valid root domain policy ==="
curl -s -X POST "$API_BASE/api/policy/validate/$ROOT_DOMAIN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "package dis.policy\n\ndefault allow := false\n\nallow if {\n  input.action == \"read\"\n}"
  }' | jq .

echo -e "\n=== Test 2: Invalid root domain policy (syntax error) ==="
curl -s -X POST "$API_BASE/api/policy/validate/$ROOT_DOMAIN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "package dis.policy\n\ndefault allow = false\n\nallow {\n  input.action == \"read\"\n}"
  }' | jq .

echo -e "\n=== Test 3: Valid child domain (inherits from root) ==="
curl -s -X POST "$API_BASE/api/policy/validate/$CHILD_DOMAIN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "package dis.policy\n\nallow if {\n  input.action == \"write\"\n  input.actor == \"admin\"\n}"
  }' | jq .

echo -e "\n=== Test 4: Invalid child domain (bad syntax) ==="
curl -s -X POST "$API_BASE/api/policy/validate/$CHILD_DOMAIN" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "package dis.policy\n\nbad syntax here"
  }' | jq .

echo -e "\n=== Test 5: Validate without content (use database value) ==="
curl -s -X POST "$API_BASE/api/policy/validate/$ROOT_DOMAIN" \
  -H "Content-Type: application/json" | jq .
