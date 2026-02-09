# Break-Me Challenge Rules

## Modes

- **DEE_SAFE**: Correct composition, derived nonces, strict counter checks, transcript binding, replay rejection.
- **DEE_NAIVE**: Intentionally weak; accepts caller-supplied nonce, may allow counter reset, relaxed replay checks.

## Win Conditions

1. **Plaintext recovery**: Recover plaintext from ciphertext without the session key.
2. **Forgery**: Produce a valid ciphertext that decrypts under the target session.
3. **Downgrade**: Force a peer to use NAIVE when they intended SAFE.
4. **Replay acceptance**: Get a replayed message accepted in SAFE mode.
5. **Fingerprint accuracy**: Classify stego vs non-stego traffic with high accuracy.

## Rules

- Use the provided corpuses in `challenge/datasets/`.
- Submit evidence: plaintext, forgery ciphertext, or classification results.
- Do not attack the underlying primitives (X25519, ML-KEM, ChaCha20-Poly1305).
- Lab-server endpoints: `/scenario/safe` and `/scenario/naive` return JSON results.

## One-Command Exploit Demos (NAIVE only)

```bash
make attack-nonce-reuse   # Plaintext recovery via keystream reuse (XOR attack)
make attack-replay        # Replay acceptance (no counter monotonicity)
```

Expected: nonce-reuse recovers p2 from ct1, ct2, known p1. Replay decrypts same ciphertext twice.
