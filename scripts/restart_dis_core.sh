#!/usr/bin/env bash
set -euo pipefail

# Restart dis-core: build and run the binary using DIS_DB_PASS to form the DSN.
# Ensure you set DIS_DB_PASS in your environment before running this script.

# Optional: load environment variables from .env if present
if [ -f "$(dirname "$0")/../.env" ]; then
  # shellcheck disable=SC1090
  source "$(dirname "$0")/../.env"
fi

if [ -z "${DIS_DB_PASS-}" ]; then
  echo "WARNING: DIS_DB_PASS is empty. Set DIS_DB_PASS in the environment before running this script." >&2
fi

# Use safe expansion so this script doesn't fail when DIS_DB_PASS is unset.
export DIS_DB_DSN="postgres://dis_user:${DIS_DB_PASS:-}@localhost:5432/dis?sslmode=disable"

echo "Checking for existing process listening on :8080..."

# Try to find PIDs owning the listen socket on TCP/8080. Prefer lsof, fall back to ss, then pgrep.
PIDS=""
if command -v lsof >/dev/null 2>&1; then
  PIDS=$(lsof -t -iTCP:8080 -sTCP:LISTEN 2>/dev/null || true)
fi
if [ -z "$PIDS" ] && command -v ss >/dev/null 2>&1; then
  # ss output contains pid=NNN, extract all pid numbers for lines containing :8080
  PIDS=$(ss -ltnp 2>/dev/null | grep -E ":8080" | sed -n 's/.*pid=\([0-9]\+\).*/\1/p' | tr '\n' ' ')
fi
if [ -z "$PIDS" ]; then
  # Best-effort: look for a running dis-core process
  if command -v pgrep >/dev/null 2>&1; then
    PIDS=$(pgrep -f "dis-core" || true)
  fi
fi

if [ -n "$PIDS" ]; then
  echo "Found running process(es) on :8080: $PIDS"
  echo "Stopping existing process(es)..."
  for pid in $PIDS; do
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" || true
    fi
  done

  # Wait up to 10s for processes to exit
  SECONDS_WAITED=0
  while [ $SECONDS_WAITED -lt 10 ]; do
    ALIVE=""
    for pid in $PIDS; do
      if kill -0 "$pid" 2>/dev/null; then
        ALIVE=1
        break
      fi
    done
    if [ -z "$ALIVE" ]; then
      break
    fi
    sleep 1
    SECONDS_WAITED=$((SECONDS_WAITED+1))
  done

  # Force kill any remaining
  for pid in $PIDS; do
    if kill -0 "$pid" 2>/dev/null; then
      echo "Process $pid did not exit; sending SIGKILL"
      kill -9 "$pid" || true
    fi
  done
else
  echo "No existing process listening on :8080 found."
fi

echo "Building dis-core..."
# script is located at dis-core/scripts — parent directory is dis-core
cd "$(dirname "$0")/.." || exit 1
go build -v ./...

echo "Starting dis-core (foreground). Use Ctrl-C to stop or run script in background.)"
./dis-core
