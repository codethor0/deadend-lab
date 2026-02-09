package drbg

import (
	"bytes"
	"testing"
)

func TestDRBGDeterministic(t *testing.T) {
	d1 := NewSeed(42)
	d2 := NewSeed(42)
	buf := make([]byte, 256)
	n1, _ := d1.Read(buf)
	b1 := append([]byte(nil), buf[:n1]...)
	n2, _ := d2.Read(buf)
	b2 := buf[:n2]
	if !bytes.Equal(b1, b2) {
		t.Errorf("same seed produced different output")
	}
}

func TestDRBGCrossProcess(t *testing.T) {
	// Verify that multiple reads produce deterministic sequence.
	d := NewSeed(42)
	out1 := make([]byte, 128)
	out2 := make([]byte, 128)
	d.Read(out1)
	d.Read(out2)
	d2 := NewSeed(42)
	got1 := make([]byte, 128)
	got2 := make([]byte, 128)
	d2.Read(got1)
	d2.Read(got2)
	if !bytes.Equal(out1, got1) || !bytes.Equal(out2, got2) {
		t.Errorf("DRBG sequence not reproducible")
	}
}
