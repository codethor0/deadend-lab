package stegopq

import (
	"bytes"
	"testing"
)

func FuzzStegoRoundtrip(f *testing.F) {
	f.Add([]byte("payload"))
	f.Fuzz(func(t *testing.T, payload []byte) {
		if len(payload) > MaxPayloadSize {
			t.Skip()
		}
		for _, c := range []Carrier{CarrierA, CarrierB, CarrierC} {
			enc, err := Encode(c, payload)
			if err != nil {
				t.Skip()
			}
			dec, err := Decode(c, enc)
			if err != nil {
				t.Fatalf("Decode carrier %v: %v", c, err)
			}
			if !bytes.Equal(dec, payload) {
				t.Errorf("roundtrip carrier %v: got %x want %x", c, dec, payload)
			}
		}
	})
}
