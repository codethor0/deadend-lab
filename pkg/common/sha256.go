package common

import (
	"crypto/sha256"
)

// HashSHA256 returns SHA-256 of data.
func HashSHA256(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}
