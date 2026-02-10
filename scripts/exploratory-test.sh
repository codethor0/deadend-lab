#!/bin/bash
# Deadend-Lab Exploratory Test Harness
# Usage: ITERATIONS=100 CONCURRENCY=50 MAX_PAYLOAD=131072 DEE_PORT=9188 ./scripts/exploratory-test.sh
set -euo pipefail

ITERATIONS="${ITERATIONS:-50}"
CONCURRENCY="${CONCURRENCY:-25}"
MAX_PAYLOAD="${MAX_PAYLOAD:-65536}"
DEE_PORT="${DEE_PORT:-9188}"
TIMEOUT_SEC="${TIMEOUT_SEC:-30}"
RESULTS_DIR="${RESULTS_DIR:-./test-results/$(date +%Y%m%d-%H%M%S)}"

mkdir -p "$RESULTS_DIR"
LOG_FILE="$RESULTS_DIR/test-run.log"
FAILURES_FILE="$RESULTS_DIR/failures.json"
METRICS_FILE="$RESULTS_DIR/metrics.json"

FORBIDDEN_PATTERNS='(k_enc|k_mac|session_key|private_key|secret_key|BEGIN.*PRIVATE KEY|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36}|xox[baprs]-|AIzaSy)'
SUSPICIOUS_HEX='[0-9a-fA-F]{128,}'

cleanup() {
  echo "Cleaning up..."
  DEE_PORT="$DEE_PORT" docker compose down -t 10 2>/dev/null || true
  echo "Results in: $RESULTS_DIR"
}
trap cleanup EXIT

log() {
  echo "[$(date '+%Y-%m-%dT%H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

log "=== EXPLORATORY TEST HARNESS ==="
log "Configuration: ITER=$ITERATIONS CONC=$CONCURRENCY MAX_PAYLOAD=$MAX_PAYLOAD PORT=$DEE_PORT"

log "[1/8] Building Docker image..."
docker build -t deadend-lab:exploratory . >> "$LOG_FILE" 2>&1

log "[2/8] Starting services..."
DEE_PORT="$DEE_PORT" docker compose down 2>/dev/null || true
DEE_PORT="$DEE_PORT" docker compose up -d >> "$LOG_FILE" 2>&1

log "[3/8] Health check..."
for i in $(seq 1 30); do
  if curl -fsS "http://localhost:${DEE_PORT}/health" >/dev/null 2>&1; then
    log "Health check: PASS"
    break
  fi
  sleep 1
done

if ! curl -fsS "http://localhost:${DEE_PORT}/health" >/dev/null 2>&1; then
  log "BLOCKER: Health check failed"
  exit 1
fi

pound() {
  local mode="$1"
  local n="$2"
  for _ in $(seq 1 "$n"); do
    curl -fsS -X POST "http://localhost:${DEE_PORT}/scenario/${mode}" --max-time "$TIMEOUT_SEC" >/dev/null 2>&1 || true
  done
}

concurrent_load_test() {
  local mode="$1"
  local concurrency="$2"
  local failed=0
  for _ in $(seq 1 "$concurrency"); do
    (curl -fsS -X POST "http://localhost:${DEE_PORT}/scenario/${mode}" --max-time "$TIMEOUT_SEC" >/dev/null 2>&1) || ((failed++)) || true
  done
  wait 2>/dev/null || true
  return 0
}

TOTAL_REQUESTS=0
FAILED_REQUESTS=0
FAILURES=()

log "[4/8] Starting exploratory test iterations..."

for iter in $(seq 1 "$ITERATIONS"); do
  log "=== ITERATION $iter/$ITERATIONS ==="

  mode=$([ $((RANDOM % 2)) -eq 0 ] && echo "safe" || echo "naive")
  if curl -fsS -X POST "http://localhost:${DEE_PORT}/scenario/${mode}" --max-time "$TIMEOUT_SEC" >/dev/null 2>&1; then
    : $((TOTAL_REQUESTS++))
  else
    : $((TOTAL_REQUESTS++))
    : $((FAILED_REQUESTS++))
    FAILURES+=("scenario_${mode}_iter_${iter}")
  fi

  if [ $((iter % 10)) -eq 0 ]; then
    log "Concurrent load test..."
    pound safe "$CONCURRENCY"
    pound naive "$CONCURRENCY"
  fi

  if [ $((iter % 25)) -eq 0 ]; then
    log "Container churn: restart"
    DEE_PORT="$DEE_PORT" docker compose restart >> "$LOG_FILE" 2>&1
    sleep 2
    if ! curl -fsS "http://localhost:${DEE_PORT}/health" >/dev/null 2>&1; then
      log "BLOCKER: Service did not recover after restart"
      exit 1
    fi
  fi

  DEE_PORT="$DEE_PORT" docker compose logs --no-color > "$RESULTS_DIR/compose.log" 2>&1 || true
  if grep -Ei "$FORBIDDEN_PATTERNS" "$RESULTS_DIR/compose.log" 2>/dev/null; then
    log "BLOCKER: Forbidden pattern in logs"
    exit 1
  fi
  if grep -E "$SUSPICIOUS_HEX" "$RESULTS_DIR/compose.log" 2>/dev/null; then
    log "WARNING: Suspicious long hex in logs"
  fi
done

log "[5/8] Attack demos..."
if ! go run ./cmd/attacks/nonce-reuse 2>> "$LOG_FILE" | grep -q 'Recovered plaintext == expected: true'; then
  FAILURES+=("attack_nonce_reuse")
fi
if ! go run ./cmd/attacks/replay 2>> "$LOG_FILE" | grep -q 'Replay accepted: true'; then
  FAILURES+=("attack_replay")
fi

log "[6/8] Generating metrics..."

if command -v jq >/dev/null 2>&1; then
  echo "{\"test_run_id\":\"$(basename "$RESULTS_DIR")\",\"timestamp\":\"$(date -Iseconds 2>/dev/null || date '+%Y-%m-%dT%H:%M:%S')\",\"configuration\":{\"iterations\":$ITERATIONS,\"concurrency\":$CONCURRENCY,\"max_payload\":$MAX_PAYLOAD,\"port\":$DEE_PORT},\"summary\":{\"total_requests\":$TOTAL_REQUESTS,\"failed_requests\":$FAILED_REQUESTS},\"failures_count\":${#FAILURES[@]},\"artifacts_location\":\"$RESULTS_DIR\"}" | jq . > "$METRICS_FILE"
else
  echo "total_requests=$TOTAL_REQUESTS failed_requests=$FAILED_REQUESTS failures=${#FAILURES[@]}" > "$METRICS_FILE"
fi

log "[7/8] Writing failure report..."

if [ ${#FAILURES[@]} -gt 0 ]; then
  if command -v jq >/dev/null 2>&1; then
    printf '%s\n' "${FAILURES[@]}" | jq -R . | jq -s . > "$FAILURES_FILE"
  else
    printf '%s\n' "${FAILURES[@]}" > "${FAILURES_FILE%.json}.txt"
  fi
else
  echo "[]" > "$FAILURES_FILE" 2>/dev/null || true
fi

log "[8/8] Test run complete"

echo ""
echo "=== TEST SUMMARY ==="
echo "Results directory: $RESULTS_DIR"
echo "Total requests: $TOTAL_REQUESTS"
echo "Failed requests: $FAILED_REQUESTS"
echo "Failure count: ${#FAILURES[@]}"
echo ""

if [ ${#FAILURES[@]} -gt 0 ]; then
  echo "FAILURES:"
  for f in "${FAILURES[@]}"; do echo "$f"; done
  exit 1
else
  echo "ALL TESTS PASSED"
  exit 0
fi
