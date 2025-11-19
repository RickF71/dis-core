#!/bin/bash
# Test QR Auth Challenge/Response Flow
# This script simulates the phone app scanning QR and completing auth

echo "📱 Simulating Phone-Side QR Auth Completion"
echo "============================================"
echo ""

# Check if challenge_id is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <challenge_id> [user_id]"
    echo ""
    echo "Example:"
    echo "  $0 550e8400-e29b-41d4-a716-446655440000 testuser"
    echo ""
    echo "Steps:"
    echo "  1. Open browser to http://localhost:5173"
    echo "  2. See QR code displayed"
    echo "  3. Extract challenge_id from browser console or QR payload"
    echo "  4. Run this script with challenge_id"
    exit 1
fi

CHALLENGE_ID="$1"
USER_ID="${2:-testuser}"

echo "Challenge ID: $CHALLENGE_ID"
echo "User ID: $USER_ID"
echo ""

# Complete the challenge
echo "🔐 Completing auth challenge..."
RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/qr-complete \
    -H "Content-Type: application/json" \
    -d "{\"challenge_id\":\"$CHALLENGE_ID\",\"user_id\":\"$USER_ID\"}")

echo "Response: $RESPONSE"
echo ""

# Check if successful
if echo "$RESPONSE" | grep -q '"ok":true'; then
    echo "✅ Challenge completed successfully!"
    echo ""
    echo "The browser should now:"
    echo "  1. Detect authenticated status"
    echo "  2. Show success message"
    echo "  3. Redirect to Finagler app"
    echo ""
    echo "Browser session is now bound to user: $USER_ID"
else
    echo "❌ Challenge completion failed"
    echo ""
    echo "Possible reasons:"
    echo "  - Challenge ID is invalid"
    echo "  - Challenge already used"
    echo "  - Challenge expired (>10 minutes old)"
    echo "  - Backend not running"
fi

echo ""
echo "============================================"
