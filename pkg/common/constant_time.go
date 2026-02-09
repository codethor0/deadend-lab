package common

import "crypto/subtle"

// EqualConstantTime returns true iff a and b are equal. Constant-time.
func EqualConstantTime(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}
