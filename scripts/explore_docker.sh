#!/usr/bin/env bash
# Deadend-Lab Docker Exploratory Test Harness
# Runs repeatable randomized endpoint pounding + restarts + log policy checks.
# Output ends with a one-page report: PASS/FAIL, failures by category, repro commands.
# Usage:
#   ITER=50 CONC=20 ./scripts/explore_docker.sh
# Requires: docker, curl, go (for attack demos)
set -euo pipefail

ITER="${ITER:-25}"
CONC="${CONC:-25}"
PORT="${DEE_PORT:-9188}"
LOG_TMP="$(mktemp -d)"
ERRLOG="${LOG_TMP}/errlog.txt"
FORBIDDEN_RE='k_enc=|k_mac=|session_key|secret|private key|BEGIN (RSA|EC|OPENSSH) PRIVATE KEY|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36}|xox[baprs]-|AIzaSy'
HEX64PLUS_RE='[0-9a-fA-F]{65,}'

record() { echo "$1" >> "$ERRLOG"; }

cleanup() {
  DEE_PORT="$PORT" docker compose down >/dev/null 2>&1 || true
  echo "Logs in: $LOG_TMP"
}
trap cleanup EXIT

: > "$ERRLOG"

echo "[1/6] Preconditions"
test -z "$(git status --porcelain)" || { echo "BLOCKER: working tree not clean"; exit 1; }
command -v docker >/dev/null || { echo "BLOCKER: docker not found"; exit 1; }
command -v curl >/dev/null || { echo "BLOCKER: curl not found"; exit 1; }

echo "[2/6] Build + start"
DEE_PORT="$PORT" docker compose down >/dev/null 2>&1 || true
docker build -t deadend-lab . || { record "CAT:build|docker build -t deadend-lab ."; exit 1; }
DEE_PORT="$PORT" docker compose up -d || { record "CAT:compose|DEE_PORT=$PORT docker compose up -d"; exit 1; }
sleep 3

echo "[3/6] Health check"
curl -fsS "http://localhost:${PORT}/health" >/dev/null || { record "CAT:health|curl -fsS http://localhost:${PORT}/health"; exit 1; }

pound() {
  mode="$1"
  for _ in $(seq 1 "$CONC"); do
    (curl -fsS -X POST "http://localhost:${PORT}/scenario/${mode}" >/dev/null) || record "CAT:scenario|curl -X POST http://localhost:${PORT}/scenario/${mode}"
  done
  wait
}

scan_logs_or_fail() {
  DEE_PORT="$PORT" docker compose logs --no-color > "${LOG_TMP}/compose.log" 2>/dev/null || true
  if grep -E -i -q "${FORBIDDEN_RE}" "${LOG_TMP}/compose.log" 2>/dev/null; then
    echo "BLOCKER: forbidden secret-like pattern found in logs"
    grep -E -i -n "${FORBIDDEN_RE}" "${LOG_TMP}/compose.log" | head -20
    exit 1
  fi
  if grep -E -q "${HEX64PLUS_RE}" "${LOG_TMP}/compose.log" 2>/dev/null; then
    echo "BLOCKER: suspicious long hex string found in logs"
    grep -E -n "${HEX64PLUS_RE}" "${LOG_TMP}/compose.log" | head -20
    exit 1
  fi
}

echo "[4/6] Exploratory loop: ITER=${ITER} CONC=${CONC} PORT=${PORT}"
for i in $(seq 1 "$ITER"); do
  echo "== Iteration $i/$ITER =="
  pound safe
  pound naive
  if [ $((i % 5)) -eq 0 ]; then
    echo "Restarting compose to simulate churn"
    DEE_PORT="$PORT" docker compose restart >/dev/null
    sleep 2
    curl -fsS "http://localhost:${PORT}/health" >/dev/null || record "CAT:health_restart|curl -fsS http://localhost:${PORT}/health (after restart iter $i)"
  fi
  scan_logs_or_fail
done

echo "[5/6] Attack demos (expected behavior)"
go run ./cmd/attacks/nonce-reuse 2>&1 | grep -q 'Recovered plaintext == expected: true' || record "CAT:attack|go run ./cmd/attacks/nonce-reuse"
go run ./cmd/attacks/replay 2>&1 | grep -q 'Replay accepted: true' || record "CAT:attack|go run ./cmd/attacks/replay"

echo "[6/6] Report"
echo "=============================================="
echo "DEADEND-LAB DOCKER EXPLORATORY HARNESS REPORT"
echo "=============================================="
if [ -s "$ERRLOG" ]; then
  echo "VERDICT: FAIL"
  echo ""
  echo "FAILURES BY CATEGORY:"
  grep -E '^CAT:' "$ERRLOG" | cut -d'|' -f1 | sort | uniq -c | while read count cat; do echo "  $cat: $count"; done
  echo ""
  echo "REPRO COMMANDS (sample):"
  grep -E '^CAT:' "$ERRLOG" | cut -d'|' -f2 | head -5 | while read -r cmd; do echo "  $cmd"; done
  exit 1
else
  echo "VERDICT: PASS"
  echo "No endpoint, health, or attack demo failures."
  echo "Log policy: no secret-like patterns or long hex in compose logs."
fi
