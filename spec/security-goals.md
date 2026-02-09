# DEE Lab Security Goals and Test Mapping

## 1. Measurable Properties

| Property | Test Type | Location |
|----------|-----------|----------|
| Encrypt/decrypt roundtrip | Unit + Fuzz | pkg/dee, tests |
| Tamper causes failure | Unit + Vector | pkg/dee |
| Replay rejected (SAFE) | Unit + Vector | pkg/dee |
| Replay accepted (NAIVE) | Unit | pkg/dee |
| Counter reset rejected (SAFE) | Unit | tests/invariants |
| Nonce derivation deterministic | Invariant | tests/invariants TestNonceUniquenessSAFE |
| stegopq decode(encode(x))=x | Unit + Fuzz | pkg/stegopq |
| Constant-time tag comparison | Unit | pkg/common |
| Transcript binding | Invariant | tests/invariants TestTranscriptBinding |
| Uniform failure (tamper/wrong key/AD) | Invariant | tests/invariants TestUniformFailure |
| Replay rejected (SAFE) | Unit + Fuzz | pkg/dee, FuzzReplayRejected |

## 2. Test Mapping

### Unit Tests

- `pkg/common`: HKDF, HMAC, transcript hash, constant-time equal.
- `pkg/dee`: Handshake, encrypt, decrypt, rekey, modes.
- `pkg/stegopq`: Encode/decode per carrier.

### Deterministic Vectors

- `tests/vectors/*.json`: Fixed seed, fixed transcripts, expected keys and ciphertexts.

### Fuzz Tests

- `FuzzDEERoundtrip`: Random messages, AD; decrypt(encrypt(m))=m.
- `FuzzStegoRoundtrip`: Random payload; decode(encode(x))=x.
- `FuzzTamper`: Tampered ciphertext fails.
- `FuzzReplayRejected`: Replay causes deterministic rejection in SAFE.
