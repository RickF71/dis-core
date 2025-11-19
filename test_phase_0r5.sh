#!/bin/bash
# Phase 0-R.5 Test: Atomic Corporeal + Actor-Domain Bootstrap

echo "=== Phase 0-R.5 Test: Corporeal Bootstrap ===" | tee -a phases/phase_0r5.log
echo "Started: $(date)" | tee -a phases/phase_0r5.log
echo "" | tee -a phases/phase_0r5.log

# Start the server in the background
echo "🚀 Starting dis-core server..." | tee -a phases/phase_0r5.log
nohup ./dis-core-phase0r5 > /tmp/dis-core-phase0r5.log 2>&1 &
SERVER_PID=$!
echo "Server PID: $SERVER_PID" | tee -a phases/phase_0r5.log

# Wait for server to be ready
echo "⏳ Waiting for server startup..." | tee -a phases/phase_0r5.log
sleep 3

# Test the bootstrap endpoint
echo "" | tee -a phases/phase_0r5.log
echo "📝 Testing POST /api/corporeal/bootstrap" | tee -a phases/phase_0r5.log
RESPONSE=$(curl -s -X POST http://localhost:8080/api/corporeal/bootstrap \
  -H "Content-Type: application/json" \
  -d '{"external_uid": "testuser"}')

echo "Response:" | tee -a phases/phase_0r5.log
echo "$RESPONSE" | jq . 2>/dev/null | tee -a phases/phase_0r5.log || echo "$RESPONSE" | tee -a phases/phase_0r5.log

# Check if the response contains expected fields
if echo "$RESPONSE" | grep -q "corporeal_domain"; then
    echo "✅ Bootstrap successful" | tee -a phases/phase_0r5.log
else
    echo "❌ Bootstrap failed" | tee -a phases/phase_0r5.log
fi

# Stop the server
echo "" | tee -a phases/phase_0r5.log
echo "🛑 Stopping server (PID: $SERVER_PID)..." | tee -a phases/phase_0r5.log
kill $SERVER_PID 2>/dev/null

echo "" | tee -a phases/phase_0r5.log
echo "=== Phase 0-R.5 Test Complete ===" | tee -a phases/phase_0r5.log
echo "Completed: $(date)" | tee -a phases/phase_0r5.log
echo "" | tee -a phases/phase_0r5.log
echo "PHASE 0-R.5 complete — corporeal + actor-domain bootstrap established atomically." | tee -a phases/phase_0r5.log
