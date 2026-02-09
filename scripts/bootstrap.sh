#!/usr/bin/env bash
set -e

echo "[bootstrap] Checking Go..."
if ! command -v go &>/dev/null; then
    echo "ERROR: Go not found. Install Go 1.22+ from https://go.dev"
    exit 1
fi

VER=$(go version 2>/dev/null | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
MAJ=$(echo "$VER" | cut -d. -f1)
MIN=$(echo "$VER" | cut -d. -f2)
if [ "$MAJ" -lt 1 ] || { [ "$MAJ" -eq 1 ] && [ "$MIN" -lt 22 ]; }; then
    echo "ERROR: Go 1.22+ required, found $VER"
    exit 1
fi

echo "[bootstrap] Installing tools..."
go install golang.org/x/tools/cmd/goimports@latest 2>/dev/null || true
if ! command -v golangci-lint &>/dev/null; then
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

cd "$(dirname "$0")/.."
ROOT=$(pwd)

echo "[bootstrap] go fmt..."
go fmt ./...

echo "[bootstrap] golangci-lint..."
golangci-lint run || echo "[bootstrap] WARN: lint had issues, continuing..."

echo "[bootstrap] go test..."
go test ./...

echo "[bootstrap] go test -race..."
go test -race ./...

echo "[bootstrap] Building binaries..."
mkdir -p bin
go build -trimpath -o bin/dee-demo ./cmd/dee-demo
go build -trimpath -o bin/corpus-gen ./cmd/corpus-gen
go build -trimpath -o bin/lab-server ./cmd/lab-server

echo "[bootstrap] Generating corpus..."
./bin/corpus-gen -out challenge/datasets

echo "[bootstrap] Docker build..."
if command -v docker &>/dev/null; then
    docker build -t deadend-lab .
    echo "[bootstrap] Docker compose up..."
    docker compose up -d
    echo "[bootstrap] Done. lab-server on port \${DEE_PORT:-8080}. Override with DEE_PORT=8081 docker compose up -d"
else
    echo "[bootstrap] Docker not found. Run manually: ./bin/lab-server"
fi
