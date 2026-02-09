package common

import (
	"bytes"
	"testing"
)

func TestExtractExpand(t *testing.T) {
	secret := []byte("secret")
	salt := []byte("salt")
	prk := Extract(secret, salt)
	if len(prk) != 32 {
		t.Errorf("Extract: want 32 bytes, got %d", len(prk))
	}
	out := Expand(prk, "info", 64)
	if len(out) != 64 {
		t.Errorf("Expand: want 64 bytes, got %d", len(out))
	}
	out2 := Expand(prk, "info", 64)
	if !bytes.Equal(out, out2) {
		t.Error("Expand should be deterministic")
	}
}

func TestEqualConstantTime(t *testing.T) {
	a := []byte{1, 2, 3}
	b := []byte{1, 2, 3}
	c := []byte{1, 2, 4}
	if !EqualConstantTime(a, b) {
		t.Error("expected equal")
	}
	if EqualConstantTime(a, c) {
		t.Error("expected not equal")
	}
}
