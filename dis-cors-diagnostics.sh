#!/usr/bin/env bash
# dis-cors-diagnostics.sh
# Gather everything needed to debug CORS + cookie auth
set -e

echo "=== ENVIRONMENT VARIABLES (DIS + Finagler) ==="
env | grep -E "DIS|PORT|ORIGIN|VITE" || echo "(none)"

echo
echo "=== ACTIVE LISTENERS (ports) ==="
lsof -nP -iTCP -sTCP:LISTEN | grep -E "node|vite|go|dis|5173|8080" || echo "(none)"

echo
echo "=== CURL TEST: Simple GET (/api/dev/users) ==="
curl -I -v http://localhost:8080/api/dev/users || true

echo
echo "=== CURL TEST: AUTH DEV (no creds) ==="
curl -I -v -X OPTIONS http://localhost:8080/auth/dev \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" || true

echo
echo "=== CURL TEST: AUTH DEV (with creds) ==="
curl -I -v -X POST http://localhost:8080/auth/dev \
  -H "Origin: http://localhost:5173" \
  -H "Content-Type: application/json" \
  --cookie "testcookie=1" || true

echo
echo "=== TRY FETCHING CORS HEADERS FROM DIS-CORE ==="
curl -s -D - http://localhost:8080/api/status -o /dev/null || true

echo
echo "=== NODE / VITE VERSION ==="
node -v || echo "node missing"
npm -v || echo "npm missing"

echo
echo "=== GIT HEAD (to ensure correct branch) ==="
git rev-parse --abbrev-ref HEAD
git rev-parse HEAD

echo
echo "=== SUMMARY ==="
echo "If any preflight request (OPTIONS) does NOT return Access-Control-Allow-Origin,"
echo "or if POST does not return Access-Control-Allow-Credentials, CORS will fail."
echo "Paste this entire output into ChatGPT for exact fix."
