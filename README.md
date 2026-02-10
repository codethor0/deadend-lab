<!-- deadend-lab:logo -->
<p align="center">
  <img src="assets/deadend-lab-logo.png" alt="deadend-lab logo" width="420" />
</p>
<!-- /deadend-lab:logo -->

# Dead-End Encryption (DEE) Lab

**RESEARCH / CTF ONLY - DO NOT USE IN PRODUCTION**

Research and challenge harness for Dead-End Encryption (DEE) and StegoPQ. This is not production crypto.

## What This Repo Is

A lab for studying session-based encryption with hybrid classical/post-quantum key agreement. DEE combines X25519 and ML-KEM (Kyber) in a fused handshake via HKDF. Session keys derive from the transcript. The protocol exposes two modes: NAIVE (intentionally weak) and SAFE (invariant-hardened). StegoPQ adds optional carrier encoding of handshake payloads for anti-fingerprinting research (see spec/stegopq.md).

## Scope

- **In scope**: Protocol implementation, invariant tests, attack demos, challenge scoring, deterministic vectors.
- **Out of scope**: Production deployment, real security guarantees, compliance validation.

## Modes

- **NAIVE**: Intentionally breakable. Caller-supplied nonce allows nonce reuse; no counter monotonicity enforces replay. Use for CTF and learning.
- **SAFE**: Nonce uniqueness, counter monotonicity, uniform failure on tamper. Resists nonce reuse and replay.

## Quick Start

```bash
./scripts/bootstrap.sh
```

Requires: Go 1.22+, Docker (optional). Bootstrap runs fmt, lint, test, build, corpus-gen, docker build, and docker compose up.

## Manual Commands

```bash
make fmt lint test test-race fuzz build
make corpus
make vectors
make docker-build docker-run
make attack-nonce-reuse
make attack-replay
```

## Demo CLI

```bash
./bin/dee-demo -mode SAFE -msg "hello"
./bin/dee-demo -mode NAIVE -msg "test"
```

## Lab Server Endpoints

- `POST /scenario/safe` - Run SAFE mode handshake + encrypt/decrypt roundtrip.
- `POST /scenario/naive` - Same for NAIVE mode.
- `GET /health` - Health check.

Response schema: `{ok, mode, version, carrier, reason_code, handshake_ms, encrypt_ms, decrypt_ms, ciphertext_len, replay_rejected, session_id_trunc}`

```bash
curl -X POST http://localhost:${DEE_PORT:-8080}/scenario/safe
curl -X POST http://localhost:${DEE_PORT:-8080}/scenario/naive
curl http://localhost:${DEE_PORT:-8080}/health
```

If port 8080 is busy, override: `DEE_PORT=9188 docker compose up -d`

## How to Break NAIVE

1. **Nonce reuse** (plaintext recovery): `make attack-nonce-reuse` - EncryptNaiveWithNonce allows caller-supplied nonce. Same nonce twice yields ct1 XOR ct2 = p1 XOR p2; known p1 recovers p2.
2. **Replay**: `make attack-replay` - NAIVE does not enforce counter monotonicity; same ciphertext decrypts multiple times.

## Why SAFE Resists

- **TestNonceUniquenessSAFE**: Every (counter, AD) yields unique nonce; no caller-supplied nonce.
- **TestCounterMonotonicityRejectReplay**: Strict counter; replay returns ErrDecrypt.
- **TestCounterMonotonicityRejectOutOfOrder**: Counter must equal next expected; out-of-order rejected.
- **TestSAFERejectsCallerNonce**: EncryptNaiveWithNonce returns ErrDecrypt in SAFE.
- **TestUniformFailure**: Tamper, wrong key, wrong AD all return identical ErrDecrypt (no oracle).

## Threat model

See [spec/threat-model.md](spec/threat-model.md) for the attacker model and assumptions.

## Specs

- spec/dee.md - Protocol, key schedule, modes
- spec/stegopq.md - Carrier encoding
- spec/threat-model.md - Attacker model
- spec/security-goals.md - Test mapping

## How to validate

Single paste-and-run from repo root (requires clean tree, Go 1.22+, Docker):

```bash
set -euo pipefail
test -z "$(git status --porcelain)" || { echo "Working tree not clean"; exit 1; }
make rc
make vectors && git diff --exit-code
make secret-scan
go test ./tests/policy/... -count=1
make verify-clean
docker build -t deadend-lab .
DEE_PORT=9188 docker compose up -d
curl -fsS http://localhost:9188/health >/dev/null
curl -fsS -X POST http://localhost:9188/scenario/safe >/dev/null
curl -fsS -X POST http://localhost:9188/scenario/naive >/dev/null
go run ./cmd/attacks/nonce-reuse
go run ./cmd/attacks/replay
DEE_PORT=9188 docker compose down
```

## Design constraints and invariants

- Nonce uniqueness (SAFE): every (counter, AD) yields a unique nonce; no caller-supplied nonce.
- Counter monotonicity: replay and out-of-order ciphertexts rejected.
- Uniform failure: tamper, wrong key, wrong AD all return identical ErrDecrypt (no oracle).
- Policy tests enforce: deterministic DRBG, handshake vectors only, no secrets in logs.

## Documentation map

Wiki disabled; docs live in `spec/` (protocol, threat model) and `challenge/` (CTF rules, scoring).

## Packages

No packages published; run from source or Docker only.

## Challenge

- challenge/break-me.md - Rules and win conditions
- challenge/scoreboard.md - Scoring rubric
- challenge/datasets/ - Generated corpuses

## Releases

See [CHANGELOG.md](CHANGELOG.md). Releases are signed; tags use SSH or GPG signing.

### Verify signatures

```bash
git log -1 --show-signature
git tag -v v0.1.0
```

### Local release-candidate run

From a clean working tree, run the full validation block (see "How to validate" above) or:

```bash
./scripts/pre-push-gate.sh
```

### Tag and push

```bash
git tag -a v0.1.1 -m "deadend-lab v0.1.1"
git push origin main --tags
```

### GitHub release notes

Include:

- **Research / CTF only** - not for production use.
- **How to reproduce validation**: Run the "How to validate" block from README.
- **How to run Docker**: `docker build -t deadend-lab .` then `DEE_PORT=9188 docker compose up -d` (override port if 8080 is busy).
- **How to break NAIVE**: `make attack-nonce-reuse` (nonce reuse), `make attack-replay` (replay).
- **Why SAFE resists**: Invariant tests (nonce uniqueness, counter monotonicity, uniform failure); policy tests (deterministic DRBG/handshake vector-only); see README "Why SAFE Resists".

### Post-release feedback

Contributions welcome: break NAIVE via demos, add attacks as `cmd/attacks/*`, add invariants/policy tests (do not weaken existing ones). File issues with repro steps and `make release-check` output.

## Security Features

### Repository Hardening Status
- Branch protection: Enabled
- CodeQL scanning: Active
- Secret scanning: Active
- Private vulnerability reporting: Enabled
- Commit attribution policy: Enforced
- Dependabot security alerts: Active (no auto-PRs)

### Reporting
- **Private vulnerability reporting:** Report sensitive findings via [GitHub advisory](https://github.com/codethor0/deadend-lab/security/advisories/new).
- **Dependabot alerts:** Go module vulnerabilities monitored weekly.
- **CodeQL scanning:** Static analysis on push, PR, and weekly schedule.
- **Secret scanning:** Detection of accidentally committed secrets.
- **Policy:** See [SECURITY.md](SECURITY.md).

## Author / Maintainer

- **Thor Thor**
- GitHub: [@codethor0](https://github.com/codethor0)
- Email: codethor@gmail.com

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution rules. Changes require `make verify` to pass. Do not weaken tests or relax security properties. License: MIT. Responsible disclosure: [SECURITY.md](SECURITY.md).

