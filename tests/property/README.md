# Property-Based Tests

This directory contains property-based tests using Go's testing/quick or gofuzz.

Properties under test:
- Nonce uniqueness: forall counters, nonces are unique
- Ciphertext indistinguishability: forall messages, ciphertexts are random-looking
- Decryption correctness: forall (key, nonce, plaintext), decrypt(encrypt(pt)) == pt
- Counter monotonicity: forall sequences, counters never decrease
