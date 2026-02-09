package stegopq

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRoundtrip(t *testing.T) {
	payloads := [][]byte{
		[]byte("hello"),
		[]byte(""),
		bytes.Repeat([]byte("x"), 100),
		{0x00, 0xff, 0x80},
	}
	for _, c := range []Carrier{CarrierA, CarrierB, CarrierC} {
		for i, p := range payloads {
			enc, err := Encode(c, p)
			if err != nil {
				t.Errorf("Carrier %v payload %d encode: %v", c, i, err)
				continue
			}
			dec, err := Decode(c, enc)
			if err != nil {
				t.Errorf("Carrier %v payload %d decode (enc=%q): %v", c, i, enc, err)
				continue
			}
			if !bytes.Equal(dec, p) {
				t.Errorf("Carrier %v payload %d roundtrip: got %x want %x", c, i, dec, p)
			}
		}
	}
}

func TestPayloadTooLarge(t *testing.T) {
	large := make([]byte, MaxPayloadSize+1)
	_, err := Encode(CarrierA, large)
	if err != ErrPayloadTooLarge {
		t.Errorf("expected ErrPayloadTooLarge, got %v", err)
	}
}
