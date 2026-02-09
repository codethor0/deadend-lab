package common

import (
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/hkdf"
)

// Extract derives a pseudorandom key from secret and salt.
func Extract(secret, salt []byte) []byte {
	return hkdf.Extract(sha256.New, secret, salt)
}

// Expand derives output from prk using info label.
func Expand(prk []byte, info string, size int) []byte {
	r := hkdf.Expand(sha256.New, prk, []byte(info))
	out := make([]byte, size)
	_, _ = io.ReadFull(r, out)
	return out
}

// ExpandLabel is a convenience for HKDF-Expand with a label.
func ExpandLabel(prk []byte, label string, size int) []byte {
	return Expand(prk, label, size)
}
