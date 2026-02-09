# Dead-End Encryption (DEE) Protocol Specification

## 1. Overview

Dead-End Encryption (DEE) is a session-based messaging protocol for research and challenge purposes. It combines classical X25519 key exchange with post-quantum ML-KEM (Kyber/NIST ML-KEM) in a hybrid handshake, fused via HKDF, to establish authenticated session keys.

**Status**: Research harness only. NOT for production cryptographic deployment.

## 2. Message Format

All messages use length-delimited binary framing:

```
[version:1][mode:1][session_id:32][counter:8][flags:2][payload_len:4][payload:N]
```

- **version** (1 byte): Protocol version. Current: 0x01.
- **mode** (1 byte): 0x01 = DEE_SAFE, 0x02 = DEE_NAIVE.
- **session_id** (32 bytes): Session identifier (SHA-256 of handshake transcript).
- **counter** (8 bytes, big-endian): Message sequence number. Monotonic for sender.
- **flags** (2 bytes): Reserved. Rekey flag bit 0.
- **payload_len** (4 bytes, big-endian): Length of payload.
- **payload**: Ciphertext (AEAD output) or handshake message.

### Handshake Message Format

- **Init** (type 0x01): `[version:1][mode:1][type:1][x25519_pub:32][kyber_pub:1184]`
- **Resp** (type 0x02): `[version:1][mode:1][type:1][x25519_pub:32][kyber_ct:1088]`

Responder encapsulates to initiator's Kyber pub; initiator decapsulates.

## 3. Key Schedule

### 3.1 Handshake Outputs

- **X25519**: Shared secret from ECDH.
- **ML-KEM**: Encapsulator (initiator) obtains `ss_pq`; decapsulator (responder) derives same `ss_pq`.
- **Fusion**: `K_raw = HKDF-Extract(transcript, X25519_ss || ss_pq)`
- **Master secret**: `K_ms = HKDF-Expand(K_raw, "dee-v1-master", 32)`

### 3.2 Domain Separation Labels

- `dee-v1-master` – Master secret.
- `dee-v1-aead-key` – AEAD encryption key (32 bytes for ChaCha20-Poly1305).
- `dee-v1-nonce-base` – Base for nonce derivation.
- `dee-v1-audit-tag-key` – Audit tag HMAC key.
- `dee-v1-rekey` – Rekey ratchet.

### 3.3 Per-Session Key Derivation

```
K_aead   = HKDF-Expand(K_ms, "dee-v1-aead-key", 32)
K_nonce  = HKDF-Expand(K_ms, "dee-v1-nonce-base", 32)
K_audit  = HKDF-Expand(K_ms, "dee-v1-audit-tag-key", 32)
K_rekey  = HKDF-Expand(K_ms, "dee-v1-rekey", 32)
```

## 4. Nonce Derivation (SAFE Mode)

Nonce is derived deterministically. Caller MUST NOT supply nonces.

```
nonce_input = session_id || transcript_hash || counter_be || hash(AD)
nonce = HMAC-SHA256(K_nonce, nonce_input)[0:12]
```

- `counter_be`: 8-byte big-endian counter.
- `hash(AD)`: SHA-256 of associated data.
- `transcript_hash`: SHA-256 of init_msg || resp_msg || mode || version.

## 5. Audit Tag (Optional, Recommended)

Before attempting AEAD decrypt:

```
audit_input = transcript_hash || header || counter_be
audit_tag = HMAC-SHA256(K_audit, audit_input)[0:16]
```

Verify audit tag matches before calling AEAD. On mismatch, return generic error.

## 6. AEAD

- Primitive: ChaCha20-Poly1305.
- Nonce: 12 bytes, derived (SAFE) or caller-supplied (NAIVE).
- Associated data: Header (version, mode, session_id, counter, flags).

## 7. Replay Protection

- **SAFE**: Strict counter monotonicity. Receive window = 1 (next expected only). Replay or out-of-order causes rejection.
- **NAIVE**: May reset counter or use weak checks; intentionally vulnerable.

## 8. Rekeying

- Trigger: Every N messages (default N=1000).
- Ratchet: `K_ms_new = HKDF-Expand(K_rekey, "dee-v1-rekey-ratchet" || counter_be, 32)`.
- Re-derive K_aead, K_nonce, K_audit, K_rekey from K_ms_new.

## 9. Error Handling

- Return generic error: `decryption failed`.
- Do not distinguish MAC failure vs decryption failure (constant-time comparison for tags).
- Log detailed errors internally; never expose to callers.

## 10. Security Goals

| ID | Goal | Description |
|----|------|-------------|
| P0 | Confidentiality | Ciphertext reveals nothing about plaintext. |
| P1 | Integrity | Tampering causes decryption failure. |
| P2 | Replay resistance | Replayed messages rejected (SAFE). |
| P3 | Nonce misuse resistance | Derived nonces prevent reuse (SAFE). |
| P4 | Transcript binding | Keys bound to handshake transcript. |

## 11. Modes

### DEE_SAFE

- Derived nonces only.
- Strict counter checks.
- Transcript binding enforced.
- Replay rejection.
- Rekeying supported.

### DEE_NAIVE

- May accept caller-supplied nonce (intentionally weak).
- May allow counter reset.
- May relax replay checks.
- For challenge/demonstration of breakage only.
