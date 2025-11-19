#!/usr/bin/env bash
# test-moar-cors.sh
# Test MOAR-CORS v1 dynamic origin checking
set -e

echo "=== MOAR-CORS v1 Testing ==="
echo

echo "Test 1: Allowed origin (localhost:5173)"
curl -I -s http://localhost:8080/api/status \
  -H "Origin: http://localhost:5173" | grep -i "access-control" || echo "  ❌ No CORS headers"

echo
echo "Test 2: Allowed origin (127.0.0.1:5173)"
curl -I -s http://localhost:8080/api/status \
  -H "Origin: http://127.0.0.1:5173" | grep -i "access-control" || echo "  ❌ No CORS headers"

echo
echo "Test 3: Disallowed origin (evil.com)"
curl -I -s http://localhost:8080/api/status \
  -H "Origin: http://evil.com" | grep -i "access-control" || echo "  ✅ Correctly blocked (no CORS headers)"

echo
echo "Test 4: No origin header (same-origin request)"
curl -I -s http://localhost:8080/api/status | grep -i "access-control" || echo "  ✅ No CORS headers (not needed)"

echo
echo "Test 5: OPTIONS preflight with allowed origin"
curl -I -s -X OPTIONS http://localhost:8080/api/status \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: GET" | grep -i "access-control" || echo "  ❌ Preflight failed"

echo
echo "Test 6: Check all expected headers present"
echo "Expected headers:"
echo "  - Access-Control-Allow-Origin: http://localhost:5173"
echo "  - Access-Control-Allow-Credentials: true"
echo "  - Access-Control-Allow-Headers: Content-Type, Authorization, X-External-User, X-Acting-As, X-Requested-With, Accept"
echo "  - Access-Control-Allow-Methods: GET,POST,OPTIONS"
echo "  - Vary: Origin"
echo
curl -I -s http://localhost:8080/api/status \
  -H "Origin: http://localhost:5173"

echo
echo "=== Environment Variable Test (DIS_ALLOWED_ORIGINS) ==="
echo "To test custom origins, run:"
echo "  export DIS_ALLOWED_ORIGINS='https://finagler.dis.example,https://app.dis.example'"
echo "  ./dis-core-moar-cors"
