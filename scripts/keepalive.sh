#!/usr/bin/env bash
set -euo pipefail

URL="${1:-https://your-render-service.onrender.com/api/v1/subscription-plans}"
INTERVAL_SECONDS="${2:-600}"
TIMEOUT_SECONDS="${3:-20}"

echo "[keepalive] Starting ping loop"
echo "[keepalive] Url=${URL} Interval=${INTERVAL_SECONDS}s Timeout=${TIMEOUT_SECONDS}s"

while true; do
  ts="$(date '+%Y-%m-%d %H:%M:%S')"
  status="$(curl -sS -o /dev/null -w '%{http_code}' --max-time "$TIMEOUT_SECONDS" "$URL" || true)"
  if [[ "$status" =~ ^2|3|4|5 ]]; then
    echo "[$ts] status=$status"
  else
    echo "[$ts] ERROR request failed"
  fi
  sleep "$INTERVAL_SECONDS"
done
