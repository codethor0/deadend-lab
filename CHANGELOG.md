# Changelog

## [v0.1.0] (research preview)

- SAFE and NAIVE lab scenarios via lab-server
- Nonce-reuse and replay attack demos against NAIVE
- Invariant tests: nonce uniqueness, counter monotonicity, uniform failure
- Policy tests: deterministic DRBG/handshake vector-only; no weakening, no auto-regenerating vectors
- Docker + compose for reproducible runs

### Safety constraints

- Deterministic DRBG and deterministic handshake are **vector-only**; never reachable from lab-server or normal flows
- Exported crypto API returns only `ErrDecrypt` on failure (no oracle)
- Lab-server `reason_code` is only `ok` or `error`; no secrets in responses
