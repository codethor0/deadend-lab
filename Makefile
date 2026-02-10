.PHONY: fmt lint test test-race test-repeat verify verify-clean release-check secret-scan rc clean fuzz build docker-build docker-run corpus vectors

clean:
	rm -rf bin dist coverage tmp

fmt:
	go fmt ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || (echo "golangci-lint not found, skipping"; exit 0)

test:
	go test ./...

test-race:
	go test -race ./...

test-repeat:
	go test -count=25 ./...
	go test -race -count=10 ./...

verify: fmt lint test test-race fuzz build vectors test-repeat

verify-clean: clean verify
	@if git rev-parse --git-dir >/dev/null 2>&1; then \
		git diff --exit-code || (echo "verify-clean: vectors or other files modified" && exit 1); \
	fi

release-check:
	./scripts/release-check.sh

rc: release-check

secret-scan:
	@./scripts/secret-scan.sh

fuzz:
	go test ./pkg/dee -fuzz=FuzzDEERoundtrip -fuzztime=10s
	go test ./pkg/dee -fuzz=FuzzTamper -fuzztime=5s
	go test ./pkg/dee -fuzz=FuzzReplayRejected -fuzztime=5s
	go test ./pkg/stegopq -fuzz=FuzzStegoRoundtrip -fuzztime=10s

build:
	mkdir -p bin
	go build -trimpath -o bin/dee-demo ./cmd/dee-demo
	go build -trimpath -o bin/corpus-gen ./cmd/corpus-gen
	go build -trimpath -o bin/lab-server ./cmd/lab-server
	go build -trimpath -o bin/vectors-gen ./cmd/vectors-gen

docker-build:
	docker build -t deadend-lab .

docker-run:
	docker compose up -d

corpus:
	go run ./cmd/corpus-gen -out challenge/datasets

vectors:
	go run ./cmd/vectors-gen -out tests/vectors/testdata

attack-nonce-reuse:
	go run ./cmd/attacks/nonce-reuse

attack-replay:
	go run ./cmd/attacks/replay
