# Scoring Rubric

| Metric | Points | Description |
|--------|--------|-------------|
| Plaintext recovery | 100 | Recover any plaintext from SAFE ciphertext |
| Forgery | 100 | Produce valid ciphertext for arbitrary plaintext |
| Downgrade | 75 | Force SAFE handshake to NAIVE |
| Replay acceptance | 50 | Replayed message accepted in SAFE |
| Fingerprint accuracy | 25 | >90% accuracy on stego vs non-stego |
| Nonce reuse break (NAIVE) | 25 | Recover plaintext from nonce-reused NAIVE pair |

## Fingerprint Test

- Generate mixed traffic via corpus-gen and stegopq.
- Train a simple classifier (regex, length, entropy).
- Report detection accuracy on held-out set.
