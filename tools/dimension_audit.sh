echo "[1] Files mentioning both 'null' and 'terra'"
echo "[2] Hard-coded spine sequences (null, terra, numen, lima, corporeal)"
echo "[3] Hard-coded dimension constants around known domains (dimension, dim, Dimension)"
echo "[4] SQL migrations that insert domains or set parent relationships"
echo "[5] Dimension increment patterns (dimension+1, Dimension+1, d.Dimension+1)"
echo "Audit complete. Review hits above and use docs/DIS_DimensionAudit.md for checklist guidance."
echo "Reminder: make executable with: chmod +x tools/dimension_audit.sh"
#!/usr/bin/env sh
set -eu

# DIS Dimension Audit Script
# Searches the repository for potential hard-coded assumptions or mappings
# between null → terra and related dimension names.
# Canonical dimension mapping (for reference):
#   0: void
#   1: (time vector / root axis) [currently named "null" in code]
#   2: aether
#   3: terra
#   4: numen
#   5: lima
#   6: corporeal

# After creation, make executable:
#   chmod +x tools/dimension_audit.sh

# Output file
OUTFILE="dimension_audit.out"
: > "$OUTFILE"

# detect ripgrep (rg) or fallback to grep
if command -v rg >/dev/null 2>&1; then
    RG_BASE_ARGS="--hidden --no-heading --line-number --max-filesize 500k --color=never"
else
    RG_BASE_ARGS=""
fi

search_all_terms_in_file() {
    # Ensure args are present
    if [ "$#" -lt 2 ]; then
        return 0
    fi
    first="$1"
    shift
    if command -v rg >/dev/null 2>&1; then
        # find files matching the first term then filter for others
        rg $RG_BASE_ARGS --files-with-matches --no-messages --glob '!**/public/**' --glob '!**/data/**' --glob '!**/overlay/**' --glob '!**/terra/**' --glob '!$OUTFILE' "$first" | while IFS= read -r f; do
            ok=1
            for t in "$@"; do
                if ! rg $RG_BASE_ARGS --no-messages -q "$t" "$f"; then
                    ok=0
                    break
                fi
            done
            if [ "$ok" -eq 1 ]; then
                rg $RG_BASE_ARGS --glob '!**/public/**' --glob '!**/data/**' --glob '!**/overlay/**' --glob '!**/terra/**' --glob '!$OUTFILE' "$first|$*" "$f" >> "$OUTFILE" || true
            fi
        done
    else
        # grep fallback: list files with first term then check others
        grep -Ril --line-number --binary-files=without-match --exclude-dir='public' --exclude-dir='data' --exclude-dir='overlay' --exclude-dir='terra' --exclude="$OUTFILE" "$first" . | while IFS= read -r f; do
            ok=1
            for t in "$@"; do
                if ! grep -qi --exclude-dir='public' --exclude-dir='data' --exclude-dir='overlay' --exclude-dir='terra' --exclude="$OUTFILE" "$t" "$f"; then
                    ok=0
                    break
                fi
            done
            if [ "$ok" -eq 1 ]; then
                grep -n --color=never -E "$first|$*" "$f" >> "$OUTFILE" || true
            fi
        done
    fi
}

echo "== DIS Dimension Audit ==" >> "$OUTFILE"
echo "Searching for legacy null → terra / dimension assumptions..." >> "$OUTFILE"

echo >> "$OUTFILE"
echo "[1] Files mentioning both 'null' and 'terra'" >> "$OUTFILE"
search_all_terms_in_file null terra || true

echo >> "$OUTFILE"
echo "[2] Hard-coded spine sequences (null, terra, numen, lima, corporeal)" >> "$OUTFILE"
# Look for files containing at least two of the canonical spine names, prefer sequences
search_all_terms_in_file null terra numen lima corporeal || true
# Also show common sequences single-line
if command -v rg >/dev/null 2>&1; then
    rg $RG_BASE_ARGS --glob '!**/public/**' --glob '!**/data/**' --glob '!**/overlay/**' --glob '!**/terra/**' --glob '!$OUTFILE' "null.*terra|terra.*null|null.*numen|terra.*numen|numen.*lima|lima.*corporeal" >> "$OUTFILE" || true
else
    grep -Rin --color=never -E "null.*terra|terra.*null|null.*numen|terra.*numen|numen.*lima|lima.*corporeal" --exclude-dir='public' --exclude-dir='data' --exclude-dir='overlay' --exclude-dir='terra' --exclude="$OUTFILE" . >> "$OUTFILE" || true
fi

echo >> "$OUTFILE"
echo "[3] Hard-coded dimension constants around known domains (dimension, dim, Dimension)" >> "$OUTFILE"
if command -v rg >/dev/null 2>&1; then
    rg $RG_BASE_ARGS --glob '!**/public/**' --glob '!**/data/**' --glob '!**/overlay/**' --glob '!**/terra/**' --glob '!$OUTFILE' "\bdimension\b|\bdim\b|\bDimension\b" >> "$OUTFILE" || true
else
    grep -Rin --color=never -E "\bdimension\b|\bdim\b|\bDimension\b" --exclude-dir='public' --exclude-dir='data' --exclude-dir='overlay' --exclude-dir='terra' --exclude="$OUTFILE" . >> "$OUTFILE" || true
fi

echo >> "$OUTFILE"
echo "[4] SQL migrations that insert domains or set parent relationships" >> "$OUTFILE"
if command -v rg >/dev/null 2>&1; then
    rg $RG_BASE_ARGS --glob '!**/public/**' --glob '!**/data/**' --glob '!**/overlay/**' --glob '!**/terra/**' --glob '!$OUTFILE' "INSERT INTO domains|INSERT INTO domain|parent_id|parent_domain|INSERT INTO domains\s*\(" >> "$OUTFILE" || true
else
    grep -Rin --color=never -E "INSERT INTO domains|INSERT INTO domain|parent_id|parent_domain|INSERT INTO domains\s*\(" --exclude-dir='public' --exclude-dir='data' --exclude-dir='overlay' --exclude-dir='terra' --exclude="$OUTFILE" . >> "$OUTFILE" || true
fi

echo >> "$OUTFILE"
echo "[5] Dimension increment patterns (dimension+1, Dimension+1, d.Dimension+1)" >> "$OUTFILE"
if command -v rg >/dev/null 2>&1; then
    rg $RG_BASE_ARGS --glob '!**/public/**' --glob '!**/data/**' --glob '!**/overlay/**' --glob '!**/terra/**' --glob '!$OUTFILE' "dimension\+1|Dimension\+1|d\.Dimension\+1|dim\+1" >> "$OUTFILE" || true
else
    grep -Rin --color=never -E "dimension\+1|Dimension\+1|d\.Dimension\+1|dim\+1" --exclude-dir='public' --exclude-dir='data' --exclude-dir='overlay' --exclude-dir='terra' --exclude="$OUTFILE" . >> "$OUTFILE" || true
fi

echo >> "$OUTFILE"
echo "Audit complete. Review hits in $OUTFILE and use docs/DIS_DimensionAudit.md for checklist guidance." >> "$OUTFILE"

echo >> "$OUTFILE"
echo "Reminder: make executable with: chmod +x tools/dimension_audit.sh" >> "$OUTFILE"
