#!/usr/bin/env bash
# Stop-the-line local gate before push. Run from repo root.
# Requires: git, make, docker, curl.
set -euo pipefail

DEE_PORT="${DEE_PORT:-9188}"
cd "$(dirname "$0")/.."

echo "== 0) Clean tree check =="
test -z "$(git status --porcelain)" || { echo "Working tree not clean"; git status --porcelain; exit 1; }

echo "== 1) Release candidate gate =="
make rc

echo "== 2) Vectors must not modify tracked files =="
make vectors
git diff --exit-code

echo "== 3) Secret-pattern scan (belt-and-suspenders) =="
make secret-scan

echo "== 4) Policy tests (tracked-files-only enforcement) =="
go test ./tests/policy/... -count=1

echo "== 5) Re-run full correctness gate =="
make verify-clean

echo "== 6) Docker build + compose up =="
docker build -t deadend-lab .
DEE_PORT="$DEE_PORT" docker compose up -d

echo "== 7) Live smoke: health + SAFE + NAIVE =="
curl -fsS "http://localhost:${DEE_PORT}/health" >/dev/null
curl -fsS -X POST "http://localhost:${DEE_PORT}/scenario/safe" >/dev/null
curl -fsS -X POST "http://localhost:${DEE_PORT}/scenario/naive" >/dev/null
echo "OK: endpoints responded"

echo "== 8) Deterministic attack demos =="
go run ./cmd/attacks/nonce-reuse
go run ./cmd/attacks/replay

echo "OK: RC gate passed end-to-end"
