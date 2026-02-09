#!/usr/bin/env bash
# Release-candidate gate: verify-clean + docker build + live smoke + attack demos.
# Requires: bash, go, docker, curl. Bounded and deterministic.
set -euo pipefail

DEE_PORT="${DEE_PORT:-9188}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

curl_retry() {
	local url="$1"
	local method="${2:-GET}"
	local i
	for i in 1 2 3 4 5; do
		if curl -fsS -X "$method" "$url" >/dev/null 2>&1; then
			return 0
		fi
		sleep 1
	done
	echo "release-check: curl failed after 5 attempts: $method $url"
	return 1
}

echo "== 1) verify-clean =="
make verify-clean

echo "== 1b) secret-scan =="
make secret-scan

echo "== 2) Docker build =="
docker build -t deadend-lab .

echo "== 3) Compose up =="
DEE_PORT="$DEE_PORT" docker compose up -d

echo "== 4) Live smoke (health + SAFE + NAIVE) =="
curl_retry "http://localhost:${DEE_PORT}/health"
curl_retry "http://localhost:${DEE_PORT}/scenario/safe" POST
curl_retry "http://localhost:${DEE_PORT}/scenario/naive" POST
echo "OK: endpoints responded"

echo "== 5) Attack demos (deterministic output) =="
go run ./cmd/attacks/nonce-reuse
go run ./cmd/attacks/replay

echo "== 6) Compose down =="
DEE_PORT="$DEE_PORT" docker compose down 2>/dev/null || true

echo "== release-check OK =="
