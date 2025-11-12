#!/usr/bin/env bash
# -------------------------------------------------------------------
# DIS Rego Editor Hierarchy Validation Test
# Uses domain IDs directly (no name lookup)
# -------------------------------------------------------------------
set -euo pipefail

API="http://localhost:8080/api/policy"
JSON='Content-Type: application/json'

# Domain IDs
NULL_ID="00000000-0000-0000-0000-000000000000"
SIMULA_ID="77de077e-3099-4393-b99b-5724451003d7"

echo "=== DIS Rego Hierarchy Test ==="
echo "Root: null ($NULL_ID)"
echo "Child: simula ($SIMULA_ID)"
echo

# 1️⃣ Get effective policy (should include parent+child)
echo "[1] GET effective bundle for simula..."
curl -s "$API/get/$SIMULA_ID?mode=effective" | jq '. | {len: (length), preview: (.[:120])}' || echo "Effective fetch failed"
echo

# 2️⃣ Validate a correct Rego (should succeed)
echo "[2] Validate valid Rego (pass expected)..."
curl -s -X POST "$API/validate/$SIMULA_ID" \
  -H "$JSON" \
  -d '{"rego":"package dis.simula\nallow if { true }"}' | jq
echo

# 3️⃣ Validate invalid Rego (should fail + return hints)
echo "[3] Validate invalid Rego (fail expected)..."
curl -s -X POST "$API/validate/$SIMULA_ID" \
  -H "$JSON" \
  -d '{"rego":"package dis.simula\nallow { true }"}' | jq
echo

# 4️⃣ Save through parent (null handles simula save)
echo "[4] Test save to parent (null handles child)..."
curl -s -X POST "$API/save/$NULL_ID" \
  -H "$JSON" \
  -d "{\"domainId\":\"$SIMULA_ID\",\"rego\":\"package dis.simula\nallow if { input.action == \\\"domain.simulation.run.v1\\\" }\"}" | jq
echo

# 5️⃣ Validate missing parent (simulate by sending child only)
echo "[5] Validate child bundle without parent (should show parent hint)..."
curl -s -X POST "$API/validate/$SIMULA_ID" \
  -H "$JSON" \
  -d '{"rego":"import data.dis.null.allow as root_allow\nallow if { root_allow }"}' | jq
echo

curl -s "$API/get/$SIMULA_ID/json?mode=effective" | jq

# 6️⃣ Wrap-up
echo "=== Test complete ==="
echo "Expected outcomes:"
echo "  [1] returns bundle preview"
echo "  [2] success:true"
echo "  [3] success:false with hints[]"
echo "  [4] success:true via parent save"
echo "  [5] success:false with 'Parent policy may not be loaded' hint"

