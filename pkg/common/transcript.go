package common

import (
	"crypto/sha256"
)

// TranscriptHash computes SHA-256 of concatenated inputs.
func TranscriptHash(inputs ...[]byte) []byte {
	h := sha256.New()
	for _, in := range inputs {
		h.Write(in)
	}
	return h.Sum(nil)
}
