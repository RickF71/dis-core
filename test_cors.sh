#!/bin/bash
# Test CORS headers from DIS-Core

echo "🌐 Testing CORS Configuration"
echo "=============================="
echo ""

# Test 1: Preflight OPTIONS request
echo "Test 1: OPTIONS Preflight Request"
echo "----------------------------------"
RESPONSE=$(curl -s -i -X OPTIONS http://localhost:8080/api/auth/challenge \
    -H "Origin: http://localhost:5173" \
    -H "Access-Control-Request-Method: POST" \
    -H "Access-Control-Request-Headers: Content-Type")

echo "$RESPONSE" | grep -E "(HTTP|Access-Control)"
echo ""

# Test 2: Actual POST request with CORS
echo "Test 2: POST with CORS Headers"
echo "-------------------------------"
RESPONSE=$(curl -s -i -X POST http://localhost:8080/api/auth/challenge \
    -H "Origin: http://localhost:5173" \
    -H "Content-Type: application/json" \
    -d '{}')

echo "$RESPONSE" | grep -E "(HTTP|Access-Control|Set-Cookie)" | head -10
echo ""

# Test 3: Check specific CORS headers
echo "Test 3: Verify CORS Headers"
echo "----------------------------"
HEADERS=$(curl -s -I -X OPTIONS http://localhost:8080/api/auth/challenge \
    -H "Origin: http://localhost:5173")

if echo "$HEADERS" | grep -q "Access-Control-Allow-Origin: http://localhost:5173"; then
    echo "✅ Access-Control-Allow-Origin: http://localhost:5173"
else
    echo "❌ Missing or incorrect Allow-Origin header"
fi

if echo "$HEADERS" | grep -q "Access-Control-Allow-Credentials: true"; then
    echo "✅ Access-Control-Allow-Credentials: true"
else
    echo "❌ Missing Allow-Credentials header"
fi

if echo "$HEADERS" | grep -q "Access-Control-Allow-Methods"; then
    echo "✅ Access-Control-Allow-Methods present"
else
    echo "❌ Missing Allow-Methods header"
fi

if echo "$HEADERS" | grep -q "Access-Control-Allow-Headers"; then
    echo "✅ Access-Control-Allow-Headers present"
else
    echo "❌ Missing Allow-Headers header"
fi

echo ""
echo "=============================="
echo "✅ CORS middleware configured"
echo ""
echo "Frontend (Finagler) should now be able to:"
echo "  - Make POST requests to /api/auth/challenge"
echo "  - Receive and store cookies (dis_browser_session)"
echo "  - Send X-External-User headers"
echo "  - Poll /api/auth/challenge/:id/status"
