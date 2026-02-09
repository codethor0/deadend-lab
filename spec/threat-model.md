# DEE Lab Threat Model

## 1. Attacker Model

### In Scope

- **Passive observer**: Sees ciphertexts, handshake messages, traffic patterns.
- **Active network attacker**: Can inject, drop, replay, reorder messages.
- **Nonce misuse**: Attacker who can trigger nonce reuse (NAIVE mode).
- **Replay attacker**: Sends previously observed messages.
- **Downgrade attempt**: Tries to force NAIVE or weaker config.
- **Fingerprint classifier**: Simple statistical/regex-based detectors for stego traffic.

### Out of Scope

- **Global traffic analysis**: Nation-state level metadata correlation.
- **Side-channel attacks**: Timing, power, EM.
- **Implementation bugs**: Buffer overflows, memory safety (handled by language).
- **Quantum attacker**: Assumed mitigated by hybrid PQ; not primary focus of this harness.

## 2. What We Defend Against

| Threat | SAFE | NAIVE |
|--------|------|-------|
| Ciphertext-only confidentiality | Yes | Yes (until nonce misuse) |
| Message tampering | Yes | Yes |
| Replay | Yes | No |
| Nonce reuse | Yes (structural prevention) | No |
| Downgrade to NAIVE | Mitigated by transcript binding | N/A |
| Basic stego detection | Reduced via carriers | Reduced |

## 3. What We Do Not Defend

- Advanced traffic analysis.
- Compromise of long-term keys (no forward secrecy guarantee beyond session).
- Denial of service (resource exhaustion).
- Metadata leakage (message lengths, timing).
