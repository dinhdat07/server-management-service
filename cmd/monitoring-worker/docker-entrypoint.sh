#!/bin/sh
set -e

SIMULATOR_HOST="${SIMULATOR_HOST:-simulator}"
SIMULATOR_SUBNET="${SIMULATOR_SUBNET:-10.1.0.0/16}"
SIMULATOR_IP="$(getent hosts "$SIMULATOR_HOST" | awk '{print $1}' | head -n 1)"

if [ -n "$SIMULATOR_IP" ]; then
  ip route add "$SIMULATOR_SUBNET" via "$SIMULATOR_IP" 2>/dev/null || true
  echo "Route ready: $SIMULATOR_SUBNET via $SIMULATOR_IP"
else
  echo "WARN: simulator host '$SIMULATOR_HOST' could not be resolved"
fi

exec ./monitoring-worker
