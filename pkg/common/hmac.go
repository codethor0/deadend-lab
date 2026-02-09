package common

import (
	"crypto/hmac"
	"crypto/sha256"
)

// HMAC256 computes HMAC-SHA256(key, data).
func HMAC256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// HMAC256Truncate returns first n bytes of HMAC-SHA256.
func HMAC256Truncate(key, data []byte, n int) []byte {
	out := HMAC256(key, data)
	if n > len(out) {
		n = len(out)
	}
	return out[:n]
}
