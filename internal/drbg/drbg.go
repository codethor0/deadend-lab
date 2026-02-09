package drbg

import (
	"crypto/sha256"
	"encoding/binary"
	"io"
)

// DRBG is a deterministic random byte generator. Same seed produces identical
// output across all runs and machines. Used only for vector generation.
type DRBG struct {
	seed    []byte
	counter uint64
	buf     []byte
	off     int
}

// New returns a DRBG seeded with the given bytes. Seed is copied.
func New(seed []byte) *DRBG {
	s := make([]byte, len(seed))
	copy(s, seed)
	return &DRBG{seed: s}
}

// NewSeed creates a DRBG from a single int64 seed (big-endian bytes).
func NewSeed(seed int64) *DRBG {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(seed))
	return New(b)
}

// Read implements io.Reader. Fills p with deterministic bytes from SHA256 stream.
func (d *DRBG) Read(p []byte) (int, error) {
	n := 0
	for len(p) > 0 {
		if d.off >= len(d.buf) {
			ctr := make([]byte, 8)
			binary.BigEndian.PutUint64(ctr, d.counter)
			h := sha256.Sum256(append(append([]byte(nil), d.seed...), ctr...))
			d.buf = h[:]
			d.off = 0
			d.counter++
		}
		copied := copy(p, d.buf[d.off:])
		p = p[copied:]
		d.off += copied
		n += copied
	}
	return n, nil
}

// MustCompile ensures DRBG implements io.Reader.
var _ io.Reader = (*DRBG)(nil)
