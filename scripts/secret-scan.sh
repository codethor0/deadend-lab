#!/usr/bin/env bash
# Belt-and-suspenders: scan for obvious secret patterns in tracked files.
# Excludes tests/policy (forbidden list) and scripts/secret-scan.sh (regex patterns).
set -euo pipefail

if ! git rev-parse --git-dir >/dev/null 2>&1; then
	echo "secret-scan: not a git repo, skipping"
	exit 0
fi

EXCLUDE=(':!tests/policy/' ':!scripts/secret-scan.sh')

fail=0
if git grep -nE '(BEGIN (RSA|EC|OPENSSH) PRIVATE KEY|AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36}|xox[baprs]-)' -- "${EXCLUDE[@]}" 2>/dev/null; then
	echo "secret-scan: potential secret patterns found"
	fail=1
fi
if git grep -nE '(private_key|shared_secret|session_key|k_enc|k_mac|seed=|handshake_secret)' -- "${EXCLUDE[@]}" 2>/dev/null; then
	echo "secret-scan: potential secret-like strings found"
	fail=1
fi
if [ "$fail" -eq 1 ]; then
	exit 1
fi
echo "OK: no obvious secret-like patterns"
