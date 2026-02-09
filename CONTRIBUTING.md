# Contributing

## Verified commits

Commits that show "Verified" on GitHub indicate the author cryptographically signed them. Options:

- **GitHub web UI**: Edits via github.com are signed automatically.
- **SSH signing** (recommended): Upload an SSH public key as a Signing Key in GitHub Settings (not just auth), then:
  ```bash
  git config --global user.name "Thor Thor"
  git config --global user.email "codethor@gmail.com"
  PUBKEY="${HOME}/.ssh/id_ed25519.pub"
  test -f "$PUBKEY" || { echo "Missing $PUBKEY. Create/reuse an SSH keypair outside the repo."; exit 1; }
  git config --global gpg.format ssh
  git config --global commit.gpgsign true
  git config --global user.signingkey "$PUBKEY"
  ```
- **GPG signing**: Configure GPG and add your public key to GitHub.

No private key material or key-generation scripts belong in this repo.

## Requirements

- **`make verify` must pass** before any PR is merged.
- Do **not** weaken tests or relax validation.
- Do **not** add self-updating or auto-regenerating behavior for vector/golden files.
- Deterministic vectors must remain deterministic by construction.

## Workflow

1. Fork and branch from `main`.
2. Run `make verify` locally (includes fmt, lint, test, test-race, fuzz, build, vectors, test-repeat).
3. Ensure `make vectors` does not modify committed vector files (byte-for-byte stable).
4. Submit PR. CI must pass.

## Publish checklist

From a clean working tree, run before pushing:

```bash
./scripts/pre-push-gate.sh
```

To confirm commit signing:

```bash
git config --global --get gpg.format
git config --global --get commit.gpgsign
git config --global --get user.signingkey
```

Then: `git push -u origin main`, tag v0.1.0, push tag, create GitHub Release with Docker/break-NAIVE/SAFE-resists notes.

## Policy

- Vector generation uses `internal/drbg` and deterministic handshake paths only. See `tests/policy/` for import and symbol guards.
- Lab-server and production handshake paths must never use deterministic APIs.
- All crypto API failures must surface as generic `ErrDecrypt` (no oracle leakage).
